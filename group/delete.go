package group

import (
	"context"
	"fmt"

	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
	"github.com/spf13/cobra"
)

// GroupDeleteCmd Deletes a Group
var GroupDeleteCmd = &cobra.Command{
	Use:   "group",
	Short: "Deletes a Passbolt Group",
	Long:  `Deletes a Passbolt Group`,
	RunE:  GroupDelete,
}

func GroupDelete(cmd *cobra.Command, args []string) error {
	id, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("no ID to Delete Provided")
	}

	return util.WithClient(cmd, func(ctx context.Context, client *api.Client) error {
		if err := client.DeleteGroup(ctx, id); err != nil {
			return fmt.Errorf("deleting Group: %w", err)
		}
		return nil
	})
}
