package group

import (
	"context"
	"fmt"

	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
)

// Filters the slice groups by invoke CEL program for each group
func filterGroups(groups *[]api.Group, celCmd string, ctx context.Context) ([]api.Group, error) {
	if celCmd == "" {
		return *groups, nil
	}

	program, err := util.InitCELProgram(celCmd, groupCelEnvOptions...)
	if err != nil {
		return nil, err
	}

	filteredGroups := []api.Group{}
	for _, group := range *groups {
		val, _, err := (*program).ContextEval(ctx, groupCelEvalMap(group))
		if err != nil {
			return nil, err
		}

		if val.Value() == true {
			filteredGroups = append(filteredGroups, group)
		}
	}

	if len(filteredGroups) == 0 {
		return nil, fmt.Errorf("no such groups found with filter %v", celCmd)
	}

	return filteredGroups, nil
}
