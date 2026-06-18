package group

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Required flags are enforced by Cobra's ValidateRequiredFlags, which runs
// before RunE. Guards against a dropped MarkFlagRequired call.
func TestGroupRequiredFlags(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
		want []string
	}{
		{"create", GroupCreateCmd, []string{"name", "manager"}},
		{"get", GroupGetCmd, []string{"id"}},
		{"update", GroupUpdateCmd, []string{"id"}},
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
