package user

import "testing"

func TestUserColumnResolver_AcceptsCanonicalAndAliases(t *testing.T) {
	// Each (input → canonical) pair models how --column / --filter inputs
	// should normalize. Covers snake_case, PascalCase, lowercased-concat,
	// and case-insensitive variants for fields where the issue is visible.
	cases := []struct {
		input string
		want  string
	}{
		{"id", "id"},
		{"ID", "id"},
		{"username", "username"},
		{"Username", "username"},
		{"first_name", "first_name"},
		{"FirstName", "first_name"},
		{"firstname", "first_name"},
		{"FIRSTNAME", "first_name"},
		{"last_name", "last_name"},
		{"LastName", "last_name"},
		{"lastname", "last_name"},
		{"role", "role"},
		{"Role", "role"},
		{"created_timestamp", "created_timestamp"},
		{"CreatedTimestamp", "created_timestamp"},
		{"createdtimestamp", "created_timestamp"},
		{"modified_timestamp", "modified_timestamp"},
		{"ModifiedTimestamp", "modified_timestamp"},
	}
	for _, c := range cases {
		got, err := userColumnResolver.Normalize(c.input)
		if err != nil {
			t.Errorf("Normalize(%q) error: %v", c.input, err)
			continue
		}
		if got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestUserColumnResolver_RejectsUnknown(t *testing.T) {
	if _, err := userColumnResolver.Normalize("nope"); err == nil {
		t.Fatal("expected error for unknown column 'nope'")
	}
}

func TestUserDefaultTableColumns_AreCanonical(t *testing.T) {
	for _, col := range userDefaultTableColumns {
		if _, ok := userColumnsByName[col]; !ok {
			t.Errorf("default table column %q is not canonical", col)
		}
	}
}
