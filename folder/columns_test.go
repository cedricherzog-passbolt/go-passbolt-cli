package folder

import "testing"

func TestFolderColumnResolver_AcceptsCanonicalAndAliases(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"id", "id"},
		{"ID", "id"},
		{"folder_parent_id", "folder_parent_id"},
		{"FolderParentID", "folder_parent_id"},
		{"folderparentid", "folder_parent_id"},
		{"name", "name"},
		{"Name", "name"},
		{"created_timestamp", "created_timestamp"},
		{"CreatedTimestamp", "created_timestamp"},
		{"modified_timestamp", "modified_timestamp"},
		{"ModifiedTimestamp", "modified_timestamp"},
		{"personal", "personal"},
		{"Personal", "personal"},
	}
	for _, c := range cases {
		got, err := folderColumnResolver.Normalize(c.input)
		if err != nil {
			t.Errorf("Normalize(%q) error: %v", c.input, err)
			continue
		}
		if got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestFolderColumnResolver_RejectsUnknown(t *testing.T) {
	if _, err := folderColumnResolver.Normalize("nope"); err == nil {
		t.Fatal("expected error for unknown column 'nope'")
	}
}

func TestFolderDefaultTableColumns_AreCanonical(t *testing.T) {
	for _, col := range folderDefaultTableColumns {
		if _, ok := folderColumnsByName[col]; !ok {
			t.Errorf("default table column %q is not canonical", col)
		}
	}
}
