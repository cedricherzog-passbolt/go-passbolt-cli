package util

import (
	"strings"
	"testing"
	"time"

	"github.com/passbolt/go-passbolt/api"
)

// PrintPermissionTable must reject an unknown column with a clear error rather
// than silently rendering a blank cell.
func TestPrintPermissionTable_UnknownColumn(t *testing.T) {
	// One row so the per-column validation loop runs.
	err := PrintPermissionTable([]string{"NoSuchColumn"}, []api.Permission{{}})
	if err == nil {
		t.Fatal("expected error for unknown column")
	}
	if !strings.Contains(err.Error(), "unknown Column") {
		t.Errorf("error %q should mention 'unknown Column'", err.Error())
	}
}

func TestPermissionCell(t *testing.T) {
	created := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	modified := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	p := api.Permission{
		ID:            "perm-id",
		ACO:           "Resource",
		ACOForeignKey: "aco-fk",
		ARO:           "User",
		AROForeignKey: "aro-fk",
		Type:          15,
		Created:       &api.Time{Time: created},
		Modified:      &api.Time{Time: modified},
	}

	cases := []struct {
		column string
		want   string
	}{
		{"id", "perm-id"},
		{"ID", "perm-id"}, // case-insensitive
		{"aco", "Resource"},
		{"AcoForeignKey", "aco-fk"},
		{"aro", "User"},
		{"aroforeignkey", "aro-fk"},
		{"type", "owner"}, // Type 15 renders as its human-readable name

		{"createdtimestamp", created.Format(time.RFC3339)},
		{"modifiedtimestamp", modified.Format(time.RFC3339)},
	}

	for _, tc := range cases {
		t.Run(tc.column, func(t *testing.T) {
			got, ok := permissionCell(p, tc.column)
			if !ok {
				t.Fatalf("permissionCell(%q) reported unknown column", tc.column)
			}
			if got != tc.want {
				t.Errorf("permissionCell(%q) = %q, want %q", tc.column, got, tc.want)
			}
		})
	}

	t.Run("unknown column", func(t *testing.T) {
		if _, ok := permissionCell(p, "bogus"); ok {
			t.Error("expected unknown column to report ok=false")
		}
	})
}

func TestPermissionTypeName(t *testing.T) {
	cases := []struct {
		code int
		want string
	}{
		{1, "read"},
		{7, "update"},
		{15, "owner"},
		{99, "99"}, // unknown codes fall back to their numeric form
	}
	for _, tc := range cases {
		if got := permissionTypeName(tc.code); got != tc.want {
			t.Errorf("permissionTypeName(%d) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestPermissionsToJSONOutput(t *testing.T) {
	created := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	modified := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	permissions := []api.Permission{
		{
			ID:            "p1",
			ACO:           "Resource",
			ACOForeignKey: "r1",
			ARO:           "User",
			AROForeignKey: "u1",
			Type:          1,
			Created:       &api.Time{Time: created},
			Modified:      &api.Time{Time: modified},
		},
	}

	out := PermissionsToJSONOutput(permissions)
	if len(out) != 1 {
		t.Fatalf("expected 1 output, got %d", len(out))
	}
	got := out[0]
	if got.ID == nil || *got.ID != "p1" {
		t.Errorf("ID mismatch: %v", got.ID)
	}
	if got.Aco == nil || *got.Aco != "Resource" {
		t.Errorf("Aco mismatch: %v", got.Aco)
	}
	if got.Type == nil || *got.Type != 1 {
		t.Errorf("Type mismatch: %v", got.Type)
	}
	if got.CreatedTimestamp == nil || !got.CreatedTimestamp.Equal(created) {
		t.Errorf("CreatedTimestamp mismatch: %v", got.CreatedTimestamp)
	}
}

func TestPermissionsToJSONOutput_Empty(t *testing.T) {
	out := PermissionsToJSONOutput(nil)
	if len(out) != 0 {
		t.Errorf("expected empty output, got %v", out)
	}
}
