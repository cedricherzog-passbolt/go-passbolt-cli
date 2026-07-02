package group

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newTestGroupListCmd builds a Cobra command shaped like GroupListCmd but with
// the persistent flags from listCmd (--json, --filter) declared locally so
// parseGroupListFlags can read them without depending on the real command tree.
func newTestGroupListCmd() *cobra.Command {
	c := &cobra.Command{Use: "group"}
	f := c.Flags()
	f.StringArrayP("user", "u", []string{}, "")
	f.StringArrayP("manager", "m", []string{}, "")
	f.StringArrayP("column", "c", groupColumns.DefaultTableColumns(), "")
	f.BoolP("json", "j", false, "")
	f.String("filter", "", "")
	return c
}

func TestParseGroupListFlags_NormalizesAliasedColumns(t *testing.T) {
	c := newTestGroupListCmd()
	if err := c.ParseFlags([]string{"--column", "Name", "--column", "ID"}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	cfg, err := parseGroupListFlags(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"name", "id"}
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

func TestParseGroupListFlags_RejectsUnknownColumn(t *testing.T) {
	c := newTestGroupListCmd()
	if err := c.ParseFlags([]string{"--column", "nope"}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if _, err := parseGroupListFlags(c); err == nil {
		t.Fatal("expected error from unknown column")
	} else if !strings.Contains(err.Error(), "unknown column") {
		t.Errorf("error %q should mention 'unknown column'", err.Error())
	}
}

func TestParseGroupListFlags_MapsFilterFlags(t *testing.T) {
	c := newTestGroupListCmd()
	args := []string{
		"--user", "u1",
		"--user", "u2",
		"--manager", "m1",
		"--json",
		"--filter", "name == 'x'",
	}
	if err := c.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	cfg, err := parseGroupListFlags(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.users) != 2 || cfg.users[0] != "u1" || cfg.users[1] != "u2" {
		t.Errorf("users = %v, want [u1 u2]", cfg.users)
	}
	if len(cfg.managers) != 1 || cfg.managers[0] != "m1" {
		t.Errorf("managers = %v, want [m1]", cfg.managers)
	}
	if !cfg.jsonOutput {
		t.Error("jsonOutput should be true")
	}
	if cfg.celFilter != "name == 'x'" {
		t.Errorf("celFilter = %q, want \"name == 'x'\"", cfg.celFilter)
	}
}
