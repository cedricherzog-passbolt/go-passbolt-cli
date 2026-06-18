package folder

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newTestFolderListCmd builds a Cobra command shaped like FolderListCmd but with
// the persistent flags from listCmd (--json, --filter) declared locally so
// parseFolderListFlags can read them without depending on the real command tree.
func newTestFolderListCmd() *cobra.Command {
	c := &cobra.Command{Use: "folder"}
	f := c.Flags()
	f.StringP("search", "s", "", "")
	f.StringArrayP("folder", "f", []string{}, "")
	f.StringArrayP("group", "g", []string{}, "")
	f.StringArrayP("column", "c", folderColumns.DefaultTableColumns(), "")
	f.BoolP("json", "j", false, "")
	f.String("filter", "", "")
	return c
}

func TestParseFolderListFlags_NormalizesAliasedColumns(t *testing.T) {
	c := newTestFolderListCmd()
	if err := c.ParseFlags([]string{"--column", "Name", "--column", "FolderParentID"}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	cfg, err := parseFolderListFlags(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"name", "folder_parent_id"}
	if len(cfg.columns) != len(want) {
		t.Fatalf("columns: got %v, want %v", cfg.columns, want)
	}
	for i := range want {
		if cfg.columns[i] != want[i] {
			t.Errorf("columns[%d] = %q, want %q", i, cfg.columns[i], want[i])
		}
	}
	if !cfg.columnsChanged {
		t.Error("columnsChanged should be true when --column is set")
	}
}

func TestParseFolderListFlags_RejectsUnknownColumn(t *testing.T) {
	c := newTestFolderListCmd()
	if err := c.ParseFlags([]string{"--column", "nope"}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if _, err := parseFolderListFlags(c); err == nil {
		t.Fatal("expected error from unknown column")
	} else if !strings.Contains(err.Error(), "unknown column") {
		t.Errorf("error %q should mention 'unknown column'", err.Error())
	}
}

func TestParseFolderListFlags_MapsFilterFlags(t *testing.T) {
	c := newTestFolderListCmd()
	args := []string{
		"--search", "myfolder",
		"--folder", "p1",
		"--folder", "p2",
		"--json",
		"--filter", "name == 'x'",
	}
	if err := c.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	cfg, err := parseFolderListFlags(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.search != "myfolder" {
		t.Errorf("search = %q, want myfolder", cfg.search)
	}
	if len(cfg.parentFolders) != 2 || cfg.parentFolders[0] != "p1" || cfg.parentFolders[1] != "p2" {
		t.Errorf("parentFolders = %v, want [p1 p2]", cfg.parentFolders)
	}
	if !cfg.jsonOutput {
		t.Error("jsonOutput should be true")
	}
	if cfg.celFilter != "name == 'x'" {
		t.Errorf("celFilter = %q, want \"name == 'x'\"", cfg.celFilter)
	}
}
