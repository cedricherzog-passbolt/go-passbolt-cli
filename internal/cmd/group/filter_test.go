package group

import (
	"context"
	"strings"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// groupColumns.Filter runs CEL over api.Group, including the int user_count
// column. Fixtures set Created/Modified (deref'd by the timestamp columns during
// eval) to avoid a panic.
func groupFilterFixtures() []api.Group {
	return []api.Group{
		{ID: "1", Name: "alpha", Deleted: false, UserCount: 3, Created: &api.Time{}, Modified: &api.Time{}},
		{ID: "2", Name: "beta", Deleted: true, UserCount: 12, Created: &api.Time{}, Modified: &api.Time{}},
	}
}

func TestGroupColumns_Filter(t *testing.T) {
	ctx := context.Background()
	items := groupFilterFixtures()

	cases := []struct {
		name    string
		expr    string
		wantIDs []string
	}{
		{"equality on name", `name == "alpha"`, []string{"1"}},
		{"contains on name", `name.contains("lph")`, []string{"1"}},
		{"int comparison user_count", `user_count > 5`, []string{"2"}},
		{"bool deleted", `deleted`, []string{"2"}},
		{"AND", `user_count > 5 && deleted`, []string{"2"}},
		{"OR", `name == "alpha" || name == "beta"`, []string{"1", "2"}},
		{"PascalCase alias", `UserCount == 3`, []string{"1"}},
		{"all match", `id != ""`, []string{"1", "2"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := groupColumns.Filter(ctx, items, c.expr)
			if err != nil {
				t.Fatalf("Filter(%q) error: %v", c.expr, err)
			}
			assertGroupIDs(t, c.expr, got, c.wantIDs)
		})
	}
}

func TestGroupColumns_Filter_NoMatchError(t *testing.T) {
	_, err := groupColumns.Filter(context.Background(), groupFilterFixtures(), `name == "nope"`)
	if err == nil {
		t.Fatal("expected error when no groups match")
	}
	if !strings.Contains(err.Error(), "no such groups found with filter") {
		t.Errorf("error %q should name the groups entity", err.Error())
	}
}

func assertGroupIDs(t *testing.T, expr string, got []api.Group, wantIDs []string) {
	t.Helper()
	if len(got) != len(wantIDs) {
		t.Fatalf("Filter(%q) returned %d groups, want %d", expr, len(got), len(wantIDs))
	}
	for i, want := range wantIDs {
		if got[i].ID != want {
			t.Errorf("Filter(%q)[%d].id = %q, want %q", expr, i, got[i].ID, want)
		}
	}
}
