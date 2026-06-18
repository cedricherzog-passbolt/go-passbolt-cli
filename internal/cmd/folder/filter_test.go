package folder

import (
	"context"
	"strings"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// folderColumns.Filter runs CEL over api.Folder. Fixtures set Created/Modified
// (deref'd by the timestamp columns during eval) to avoid a panic.
func folderFilterFixtures() []api.Folder {
	return []api.Folder{
		{ID: "1", Name: "alpha", FolderParentID: "p1", Personal: false, Created: &api.Time{}, Modified: &api.Time{}},
		{ID: "2", Name: "beta", FolderParentID: "", Personal: true, Created: &api.Time{}, Modified: &api.Time{}},
	}
}

func TestFolderColumns_Filter(t *testing.T) {
	ctx := context.Background()
	items := folderFilterFixtures()

	cases := []struct {
		name    string
		expr    string
		wantIDs []string
	}{
		{"equality on name", `name == "alpha"`, []string{"1"}},
		{"contains on name", `name.contains("et")`, []string{"2"}},
		{"bool personal", `personal`, []string{"2"}},
		{"AND", `name == "beta" && personal`, []string{"2"}},
		{"OR", `name == "alpha" || name == "beta"`, []string{"1", "2"}},
		{"PascalCase alias", `Name == "alpha"`, []string{"1"}},
		{"all match", `id != ""`, []string{"1", "2"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := folderColumns.Filter(ctx, items, c.expr)
			if err != nil {
				t.Fatalf("Filter(%q) error: %v", c.expr, err)
			}
			assertFolderIDs(t, c.expr, got, c.wantIDs)
		})
	}
}

func TestFolderColumns_Filter_NoMatchError(t *testing.T) {
	_, err := folderColumns.Filter(context.Background(), folderFilterFixtures(), `name == "nope"`)
	if err == nil {
		t.Fatal("expected error when no folders match")
	}
	if !strings.Contains(err.Error(), "no such folders found with filter") {
		t.Errorf("error %q should name the folders entity", err.Error())
	}
}

func assertFolderIDs(t *testing.T, expr string, got []api.Folder, wantIDs []string) {
	t.Helper()
	if len(got) != len(wantIDs) {
		t.Fatalf("Filter(%q) returned %d folders, want %d", expr, len(got), len(wantIDs))
	}
	for i, want := range wantIDs {
		if got[i].ID != want {
			t.Errorf("Filter(%q)[%d].id = %q, want %q", expr, i, got[i].ID, want)
		}
	}
}
