package folder

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Required flags are enforced by Cobra's ValidateRequiredFlags, which runs
// before RunE. Guards against a dropped MarkFlagRequired call.
func TestFolderRequiredFlags(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
		want []string
	}{
		{"create", FolderCreateCmd, []string{"name"}},
		{"get", FolderGetCmd, []string{"id"}},
		{"permission", FolderPermissionCmd, []string{"id"}},
		{"update", FolderUpdateCmd, []string{"id", "name"}},
		{"share", FolderShareCmd, []string{"id", "type"}},
		{"move", FolderMoveCmd, []string{"id", "folderParentID"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.cmd.ValidateRequiredFlags()
			if err == nil {
				t.Fatalf("expected error for missing required flags %v", c.want)
			}
			for _, f := range c.want {
				if !strings.Contains(err.Error(), f) {
					t.Errorf("error %q should mention required flag %q", err.Error(), f)
				}
			}
		})
	}
}
