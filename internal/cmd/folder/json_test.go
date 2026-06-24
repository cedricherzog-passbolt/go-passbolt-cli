package folder

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// FolderJSONOutput.Personal must emit even when false so that
// --filter 'personal == false' isn't blocked by omitempty stripping.
func TestFolderJSONOutput_PersonalAlwaysEmitted(t *testing.T) {
	out := FolderJSONOutput{ID: ptr("f1"), Name: ptr("shared"), Personal: false}
	got, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(got), `"personal":false`) {
		t.Errorf("missing \"personal\":false in output %s", got)
	}
}

func TestPrintJSONFolders_PopulatesFields(t *testing.T) {
	folders := []api.Folder{
		{
			ID:             "f1",
			FolderParentID: "p1",
			Name:           "shared",
			Created:        &api.Time{},
			Modified:       &api.Time{},
			Personal:       false,
		},
		{
			ID:       "f2",
			Name:     "private",
			Created:  &api.Time{},
			Modified: &api.Time{},
			Personal: true,
		},
	}

	out := captureStdout(t, func() {
		if err := printJSONFolders(folders, false, nil); err != nil {
			t.Fatalf("printJSONFolders: %v", err)
		}
	})

	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v\nraw:\n%s", err, out)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 folders, got %d", len(got))
	}
	if got[0]["id"] != "f1" || got[0]["name"] != "shared" {
		t.Errorf("folder[0] = %v", got[0])
	}
	if got[0]["personal"] != false {
		t.Errorf("folder[0].personal = %v, want false", got[0]["personal"])
	}
	if got[1]["personal"] != true {
		t.Errorf("folder[1].personal = %v, want true", got[1]["personal"])
	}
}

func TestPrintJSONFolders_ColumnFilterRestrictsKeys(t *testing.T) {
	folders := []api.Folder{
		{ID: "f1", Name: "shared", FolderParentID: "p1", Created: &api.Time{}, Modified: &api.Time{}},
	}

	out := captureStdout(t, func() {
		if err := printJSONFolders(folders, true, []string{"id", "name"}); err != nil {
			t.Fatalf("printJSONFolders: %v", err)
		}
	})

	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v\nraw:\n%s", err, out)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(got))
	}
	if _, ok := got[0]["folder_parent_id"]; ok {
		t.Errorf("folder_parent_id should be filtered out, got %v", got[0])
	}
	if got[0]["name"] != "shared" {
		t.Errorf("name = %v, want shared", got[0]["name"])
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
