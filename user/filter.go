package user

import (
	"context"
	"fmt"

	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
)

// Filters the slice users by invoke CEL program for each user
func filterUsers(users *[]api.User, celCmd string, ctx context.Context) ([]api.User, error) {
	if celCmd == "" {
		return *users, nil
	}

	program, err := util.InitCELProgram(celCmd, userCelEnvOptions...)
	if err != nil {
		return nil, err
	}

	filteredUsers := []api.User{}
	for _, user := range *users {
		val, _, err := (*program).ContextEval(ctx, userCelEvalMap(user))
		if err != nil {
			return nil, err
		}

		if val.Value() == true {
			filteredUsers = append(filteredUsers, user)
		}
	}

	if len(filteredUsers) == 0 {
		return nil, fmt.Errorf("no such users found with filter %v", celCmd)
	}

	return filteredUsers, nil
}
