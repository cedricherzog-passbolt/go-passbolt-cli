package resource

import (
	"strings"
	"testing"
)

func TestResourceColumnResolver_AcceptsCanonicalAndAliases(t *testing.T) {
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
		{"username", "username"},
		{"Username", "username"},
		{"uri", "uri"},
		{"URI", "uri"},
		{"password", "password"},
		{"Password", "password"},
		{"description", "description"},
		{"Description", "description"},
		{"created_timestamp", "created_timestamp"},
		{"CreatedTimestamp", "created_timestamp"},
		{"modified_timestamp", "modified_timestamp"},
		{"ModifiedTimestamp", "modified_timestamp"},
		{"metadata", "metadata"},
		{"Metadata", "metadata"},
		{"secret", "secret"},
		{"Secret", "secret"},
		{"deleted", "deleted"},
		{"Deleted", "deleted"},
		{"expired", "expired"},
		{"Expired", "expired"},
		{"expired_at", "expired_at"},
		{"ExpiredAt", "expired_at"},
		{"expiredat", "expired_at"},
		{"resource_type_id", "resource_type_id"},
		{"ResourceTypeID", "resource_type_id"},
		{"resourcetypeid", "resource_type_id"},
	}
	for _, c := range cases {
		got, err := resourceColumns.Resolver().Normalize(c.input)
		if err != nil {
			t.Errorf("Normalize(%q) error: %v", c.input, err)
			continue
		}
		if got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestResourceColumnResolver_RejectsUnknown(t *testing.T) {
	if _, err := resourceColumns.Resolver().Normalize("nope"); err == nil {
		t.Fatal("expected error for unknown column 'nope'")
	}
}

func TestResourceDefaultTableColumns_AreCanonical(t *testing.T) {
	for _, col := range resourceColumns.DefaultTableColumns() {
		if canon, err := resourceColumns.Resolver().Normalize(col); err != nil || canon != col {
			t.Errorf("default table column %q is not canonical", col)
		}
	}
}

func TestColumnsRequireSecrets(t *testing.T) {
	cases := []struct {
		name    string
		columns []string
		want    bool
	}{
		{"no secret columns", []string{"id", "name", "username"}, false},
		{"password triggers", []string{"id", "password"}, true},
		{"description triggers", []string{"id", "description"}, true},
		{"secret map triggers", []string{"id", "secret"}, true},
		{"metadata alone does not trigger", []string{"id", "metadata"}, false},
	}
	for _, c := range cases {
		if got := resourceColumns.RequiresSecrets(c.columns); got != c.want {
			t.Errorf("%s: RequiresSecrets(%v) = %v, want %v", c.name, c.columns, got, c.want)
		}
	}
}

func TestMarshalMapForTable(t *testing.T) {
	// nil → empty string (used by tableValue for resources without metadata/secret).
	if got := marshalMapForTable(nil); got != "" {
		t.Errorf("nil map: got %q, want empty", got)
	}
	// empty map → empty string (mirrors nil; nothing meaningful to display).
	if got := marshalMapForTable(map[string]any{}); got != "" {
		t.Errorf("empty map: got %q, want empty", got)
	}
	// populated map → valid JSON whose round-trip preserves the key/value.
	in := map[string]any{"object_type": "PASSBOLT_RESOURCE_METADATA", "name": "X"}
	got := marshalMapForTable(in)
	if got == "" {
		t.Fatal("populated map: got empty string")
	}
	// Don't depend on key ordering — just verify both keys made it through.
	for _, want := range []string{`"object_type":"PASSBOLT_RESOURCE_METADATA"`, `"name":"X"`} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q missing %q", got, want)
		}
	}
}

func TestResourceSecretCelNames_ContainsCanonicalAndAliases(t *testing.T) {
	secretNames := resourceColumns.SecretCelNames()
	have := make(map[string]bool, len(secretNames))
	for _, n := range secretNames {
		have[n] = true
	}
	// Canonical names — referenced by user filter expressions in snake_case.
	for _, want := range []string{"password", "description", "secret"} {
		if !have[want] {
			t.Errorf("missing canonical CEL name %q in secret list (saw %v)", want, secretNames)
		}
	}
	// PascalCase aliases — referenced by existing user scripts.
	for _, want := range []string{"Password", "Description", "Secret"} {
		if !have[want] {
			t.Errorf("missing PascalCase CEL alias %q in secret list (saw %v)", want, secretNames)
		}
	}
}
