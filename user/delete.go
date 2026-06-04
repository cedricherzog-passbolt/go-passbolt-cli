package user

import (
	"context"
	"fmt"

	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
	"github.com/passbolt/go-passbolt/helper"
	"github.com/spf13/cobra"
)

// UserDeleteCmd Deletes a User
var UserDeleteCmd = &cobra.Command{
	Use:   "user",
	Short: "Deletes a Passbolt User",
	Long:  `Deletes a Passbolt User`,
	RunE:  UserDelete,
}

func UserDelete(cmd *cobra.Command, args []string) error {
	id, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("no ID to Delete Provided")
	}

	return util.WithClient(cmd, func(ctx context.Context, client *api.Client) error {
		if err := helper.DeleteUser(ctx, client, id); err != nil {
			return fmt.Errorf("deleting User: %w", err)
		}
		return nil
	})
}
