package user

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// userColumns.Filter runs CEL over api.User. The first_name/last_name/role
// columns deref Profile/Role, and the timestamp columns deref Created/Modified,
// so every fixture must populate them (Filter evaluates all columns per item).
func userFilterFixtures() []api.User {
	return []api.User{
		{
			ID: "1", Username: "ada@example.com", Active: true,
			Profile: &api.Profile{FirstName: "Ada", LastName: "Lovelace"},
			Role:    &api.Role{Name: "user"},
			Created: &api.Time{}, Modified: &api.Time{},
		},
		{
			ID: "2", Username: "grace@example.com", Active: false,
			Profile: &api.Profile{FirstName: "Grace", LastName: "Hopper"},
			Role:    &api.Role{Name: "admin"},
			Created: &api.Time{}, Modified: &api.Time{},
		},
	}
}

func TestUserColumns_Filter(t *testing.T) {
	ctx := context.Background()
	items := userFilterFixtures()

	cases := []struct {
		name    string
		expr    string
		wantIDs []string
	}{
		{"equality on username", `username == "ada@example.com"`, []string{"1"}},
		{"contains on first_name", `first_name.contains("ra")`, []string{"2"}},
		{"role equality", `role == "admin"`, []string{"2"}},
		{"bool active", `active`, []string{"1"}},
		{"AND", `role == "admin" && !active`, []string{"2"}},
		{"OR", `username == "ada@example.com" || username == "grace@example.com"`, []string{"1", "2"}},
		{"PascalCase alias", `FirstName == "Ada"`, []string{"1"}},
		{"all match", `id != ""`, []string{"1", "2"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := userColumns.Filter(ctx, items, c.expr)
			if err != nil {
				t.Fatalf("Filter(%q) error: %v", c.expr, err)
			}
			assertUserIDs(t, c.expr, got, c.wantIDs)
		})
	}
}

func TestUserColumns_Filter_NoMatchError(t *testing.T) {
	_, err := userColumns.Filter(context.Background(), userFilterFixtures(), `username == "nobody@example.com"`)
	if err == nil {
		t.Fatal("expected error when no users match")
	}
	if !strings.Contains(err.Error(), "no such users found with filter") {
		t.Errorf("error %q should name the users entity", err.Error())
	}
}

// Filtering must stay correct on a large input — exercises the per-item eval
// loop at scale. This is a correctness check (returned count), not a timing gate.
func TestUserColumns_Filter_LargeDataset(t *testing.T) {
	const n = 10000
	items := make([]api.User, n)
	wantActive := 0
	for i := range n {
		active := i%2 == 0
		if active {
			wantActive++
		}
		items[i] = api.User{
			ID:       fmt.Sprintf("u%d", i),
			Username: fmt.Sprintf("user%d@example.com", i),
			Active:   active,
			Profile:  &api.Profile{FirstName: "F", LastName: "L"},
			Role:     &api.Role{Name: "user"},
			Created:  &api.Time{}, Modified: &api.Time{},
		}
	}

	got, err := userColumns.Filter(context.Background(), items, `active`)
	if err != nil {
		t.Fatalf("Filter on large dataset: %v", err)
	}
	if len(got) != wantActive {
		t.Errorf("filtered %d active users, want %d", len(got), wantActive)
	}
}

func assertUserIDs(t *testing.T, expr string, got []api.User, wantIDs []string) {
	t.Helper()
	if len(got) != len(wantIDs) {
		t.Fatalf("Filter(%q) returned %d users, want %d", expr, len(got), len(wantIDs))
	}
	for i, want := range wantIDs {
		if got[i].ID != want {
			t.Errorf("Filter(%q)[%d].id = %q, want %q", expr, i, got[i].ID, want)
		}
	}
}
