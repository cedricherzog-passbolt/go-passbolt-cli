package resource

import (
	"context"
	"strings"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// resourceColumns.Filter runs CEL over the *decrypted* resource fields (name,
// username, password, ...), so these tests double as documentation of the
// supported --filter syntax for resources and verify the column CelValue/CelType
// wiring on real decryptedResource values.
//
// Filter evaluates every column's CelValue for each item, so fixtures must set
// resource.Created/Modified (deref'd by the timestamp columns) to avoid a panic.
func resourceFilterFixtures() []decryptedResource {
	return []decryptedResource{
		{
			resource:    api.Resource{ID: "1", Created: &api.Time{}, Modified: &api.Time{}, Deleted: false},
			name:        "alpha",
			username:    "ada@example.com",
			uri:         "https://alpha.example.com",
			password:    "p1",
			description: "first",
		},
		{
			resource:    api.Resource{ID: "2", Created: &api.Time{}, Modified: &api.Time{}, Deleted: true},
			name:        "beta",
			username:    "bob@example.com",
			uri:         "https://beta.example.com",
			password:    "p2",
			description: "second",
		},
	}
}

func TestResourceColumns_Filter(t *testing.T) {
	ctx := context.Background()
	items := resourceFilterFixtures()

	cases := []struct {
		name    string
		expr    string
		wantIDs []string
	}{
		{"equality on name", `name == "alpha"`, []string{"1"}},
		{"contains on username", `username.contains("bob")`, []string{"2"}},
		{"AND", `name == "alpha" && username == "ada@example.com"`, []string{"1"}},
		{"OR", `name == "alpha" || name == "beta"`, []string{"1", "2"}},
		{"bool deleted", `deleted`, []string{"2"}},
		{"secret field equality", `password == "p1"`, []string{"1"}},
		{"PascalCase alias", `Name == "beta"`, []string{"2"}},
		{"all match", `id != ""`, []string{"1", "2"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := resourceColumns.Filter(ctx, items, c.expr)
			if err != nil {
				t.Fatalf("Filter(%q) error: %v", c.expr, err)
			}
			assertResourceIDs(t, c.expr, got, c.wantIDs)
		})
	}
}

func TestResourceColumns_Filter_NoMatchError(t *testing.T) {
	_, err := resourceColumns.Filter(context.Background(), resourceFilterFixtures(), `name == "nope"`)
	if err == nil {
		t.Fatal("expected error when no resources match")
	}
	if !strings.Contains(err.Error(), "no such resources found with filter") {
		t.Errorf("error %q should name the resources entity", err.Error())
	}
}

func assertResourceIDs(t *testing.T, expr string, got []decryptedResource, wantIDs []string) {
	t.Helper()
	if len(got) != len(wantIDs) {
		t.Fatalf("Filter(%q) returned %d resources, want %d", expr, len(got), len(wantIDs))
	}
	for i, want := range wantIDs {
		if got[i].resource.ID != want {
			t.Errorf("Filter(%q)[%d].id = %q, want %q", expr, i, got[i].resource.ID, want)
		}
	}
}
