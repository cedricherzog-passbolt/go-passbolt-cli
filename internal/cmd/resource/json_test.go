package resource

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// ResourceJSONOutput's Deleted/Expired fields must emit even when false so that
// --filter 'deleted == false' isn't blocked by omitempty-style stripping.
// Regression guard mirroring the user lifecycle-field test.
func TestResourceJSONOutput_BoolFieldsAlwaysEmitted(t *testing.T) {
	out := ResourceJSONOutput{
		ID:      ptr("r1"),
		Name:    ptr("server"),
		Deleted: false,
		Expired: false,
	}
	got, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"deleted":false`, `"expired":false`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in output %s", want, got)
		}
	}
}

// printJSONResources must map decryptedResource onto ResourceJSONOutput,
// preserving the decrypted fields and resource metadata.
func TestPrintJSONResources_PopulatesFields(t *testing.T) {
	decrypted := []decryptedResource{
		{
			resource: api.Resource{
				ID:             "r1",
				FolderParentID: "f1",
				ResourceTypeID: "type-1",
				Created:        &api.Time{},
				Modified:       &api.Time{},
				Deleted:        false,
			},
			name:        "server",
			username:    "admin",
			uri:         "https://example.com",
			password:    "secret",
			description: "prod box",
		},
	}

	out := captureStdout(t, func() {
		if err := printJSONResources(decrypted, false, nil); err != nil {
			t.Fatalf("printJSONResources: %v", err)
		}
	})

	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v\nraw:\n%s", err, out)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(got))
	}
	rec := got[0]
	for field, want := range map[string]string{
		"id":       "r1",
		"name":     "server",
		"username": "admin",
		"uri":      "https://example.com",
		"password": "secret",
	} {
		if rec[field] != want {
			t.Errorf("%s = %v, want %v", field, rec[field], want)
		}
	}
}

// When the user selects a subset of columns, only those keys should survive in
// the JSON output (via util.PrintJSONColumnFiltered).
func TestPrintJSONResources_ColumnFilterRestrictsKeys(t *testing.T) {
	decrypted := []decryptedResource{
		{
			resource: api.Resource{
				ID:       "r1",
				Created:  &api.Time{},
				Modified: &api.Time{},
			},
			name:     "server",
			password: "secret",
		},
	}

	out := captureStdout(t, func() {
		if err := printJSONResources(decrypted, true, []string{"id", "name"}); err != nil {
			t.Fatalf("printJSONResources: %v", err)
		}
	})

	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v\nraw:\n%s", err, out)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(got))
	}
	if _, ok := got[0]["password"]; ok {
		t.Errorf("password should be filtered out, got %v", got[0])
	}
	if got[0]["name"] != "server" {
		t.Errorf("name = %v, want server", got[0]["name"])
	}
}

func ptr[T any](v T) *T { return &v }

// captureStdout swaps os.Stdout for a pipe, runs fn, restores stdout, and
// returns whatever fn wrote. Used to test functions that print directly.
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
