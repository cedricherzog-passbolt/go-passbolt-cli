package group

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// GroupJSONOutput's Deleted/UserCount must emit even when false/0 so that
// --filter 'deleted == false' / 'user_count == 0' isn't blocked by omitempty.
func TestGroupJSONOutput_ScalarFieldsAlwaysEmitted(t *testing.T) {
	out := GroupJSONOutput{ID: ptr("g1"), Name: ptr("eng"), Deleted: false, UserCount: 0}
	got, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"deleted":false`, `"user_count":0`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in output %s", want, got)
		}
	}
}

func TestPrintJSONGroups_PopulatesFields(t *testing.T) {
	groups := []api.Group{
		{
			ID:        "g1",
			Name:      "engineering",
			Created:   &api.Time{},
			Modified:  &api.Time{},
			Deleted:   false,
			UserCount: 5,
		},
	}

	out := captureStdout(t, func() {
		if err := printJSONGroups(groups, false, nil); err != nil {
			t.Fatalf("printJSONGroups: %v", err)
		}
	})

	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v\nraw:\n%s", err, out)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 group, got %d", len(got))
	}
	if got[0]["id"] != "g1" || got[0]["name"] != "engineering" {
		t.Errorf("group[0] = %v", got[0])
	}
	if got[0]["user_count"] != float64(5) {
		t.Errorf("user_count = %v, want 5", got[0]["user_count"])
	}
}

func TestPrintJSONGroups_ColumnFilterRestrictsKeys(t *testing.T) {
	groups := []api.Group{
		{ID: "g1", Name: "engineering", Created: &api.Time{}, Modified: &api.Time{}, UserCount: 5},
	}

	out := captureStdout(t, func() {
		if err := printJSONGroups(groups, true, []string{"id", "name"}); err != nil {
			t.Fatalf("printJSONGroups: %v", err)
		}
	})

	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v\nraw:\n%s", err, out)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 group, got %d", len(got))
	}
	if _, ok := got[0]["user_count"]; ok {
		t.Errorf("user_count should be filtered out, got %v", got[0])
	}
	if got[0]["name"] != "engineering" {
		t.Errorf("name = %v, want engineering", got[0]["name"])
	}
}

func ptr[T any](v T) *T { return &v }

// captureStdout swaps os.Stdout for a pipe, runs fn, restores stdout, and
// returns whatever fn wrote.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()
	_ = w.Close()
	<-done
	return buf.String()
}
