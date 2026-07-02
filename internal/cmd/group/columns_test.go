package group

import "testing"

func TestGroupColumnResolver_AcceptsCanonicalAndAliases(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"id", "id"},
		{"ID", "id"},
		{"name", "name"},
		{"Name", "name"},
		{"created_timestamp", "created_timestamp"},
		{"CreatedTimestamp", "created_timestamp"},
		{"createdtimestamp", "created_timestamp"},
		{"modified_timestamp", "modified_timestamp"},
		{"ModifiedTimestamp", "modified_timestamp"},
		{"deleted", "deleted"},
		{"Deleted", "deleted"},
		{"user_count", "user_count"},
		{"UserCount", "user_count"},
		{"usercount", "user_count"},
	}
	for _, c := range cases {
		got, err := groupColumns.Resolver().Normalize(c.input)
		if err != nil {
			t.Errorf("Normalize(%q) error: %v", c.input, err)
			continue
		}
		if got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestGroupColumnResolver_RejectsUnknown(t *testing.T) {
	if _, err := groupColumns.Resolver().Normalize("nope"); err == nil {
		t.Fatal("expected error for unknown column 'nope'")
	}
}

func TestGroupDefaultTableColumns_AreCanonical(t *testing.T) {
	for _, col := range groupColumns.DefaultTableColumns() {
		if canon, err := groupColumns.Resolver().Normalize(col); err != nil || canon != col {
			t.Errorf("default table column %q is not canonical", col)
		}
	}
}
