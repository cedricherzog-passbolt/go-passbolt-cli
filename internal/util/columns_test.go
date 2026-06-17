package util

import (
	"strings"
	"testing"
)

func newTestResolver() *ColumnAliasResolver {
	return NewColumnAliasResolver([]ColumnAlias{
		{Name: "id", Aliases: []string{"ID"}},
		{Name: "first_name", Aliases: []string{"FirstName", "firstname"}},
		{Name: "folder_parent_id", Aliases: []string{"FolderParentID", "folderparentid"}},
	})
}

func TestColumnAliasResolver_Normalize(t *testing.T) {
	r := newTestResolver()

	cases := []struct {
		input string
		want  string
	}{
		{"id", "id"},
		{"ID", "id"},
		{"Id", "id"},
		{"first_name", "first_name"},
		{"FirstName", "first_name"},
		{"firstname", "first_name"},
		{"FIRSTNAME", "first_name"},
		{"folder_parent_id", "folder_parent_id"},
		{"FolderParentID", "folder_parent_id"},
		{"folderparentid", "folder_parent_id"},
	}
	for _, c := range cases {
		got, err := r.Normalize(c.input)
		if err != nil {
			t.Errorf("Normalize(%q) returned error: %v", c.input, err)
			continue
		}
		if got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestColumnAliasResolver_Normalize_Unknown(t *testing.T) {
	r := newTestResolver()
	_, err := r.Normalize("nope")
	if err == nil {
		t.Fatal("expected error for unknown column")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unknown column") {
		t.Errorf("error message %q missing 'unknown column'", msg)
	}
	if !strings.Contains(msg, "nope") {
		t.Errorf("error message %q missing the offending column name", msg)
	}
	// Must list valid canonical names so the user can recover.
	for _, want := range []string{"id", "first_name", "folder_parent_id"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message %q missing valid column %q", msg, want)
		}
	}
}

func TestColumnAliasResolver_NormalizeAll(t *testing.T) {
	r := newTestResolver()

	got, err := r.NormalizeAll([]string{"ID", "FirstName", "folder_parent_id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"id", "first_name", "folder_parent_id"}
	if len(got) != len(want) {
		t.Fatalf("got %d items, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestColumnAliasResolver_NormalizeAll_StopsOnFirstError(t *testing.T) {
	r := newTestResolver()

	_, err := r.NormalizeAll([]string{"ID", "nope", "FirstName"})
	if err == nil {
		t.Fatal("expected error from invalid column in middle of slice")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error %q should reference offending input", err.Error())
	}
}

func TestColumnAliasResolver_Canonical_PreservesDeclarationOrder(t *testing.T) {
	r := newTestResolver()
	got := r.Canonical()
	want := []string{"id", "first_name", "folder_parent_id"}
	if len(got) != len(want) {
		t.Fatalf("got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q want %q", i, got[i], want[i])
		}
	}
	// Mutating the returned slice must not affect future calls.
	got[0] = "mutated"
	again := r.Canonical()
	if again[0] != "id" {
		t.Errorf("returned slice is not defensively copied: again[0] = %q", again[0])
	}
}
