package folder

import (
	"context"
	"fmt"

	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
)

// Filters the slice folders by invoke CEL program for each folder
func filterFolders(folders *[]api.Folder, celCmd string, ctx context.Context) ([]api.Folder, error) {
	if celCmd == "" {
		return *folders, nil
	}

	program, err := util.InitCELProgram(celCmd, folderCelEnvOptions...)
	if err != nil {
		return nil, err
	}

	filteredFolders := []api.Folder{}
	for _, folder := range *folders {
		val, _, err := (*program).ContextEval(ctx, folderCelEvalMap(folder))
		if err != nil {
			return nil, err
		}

		if val.Value() == true {
			filteredFolders = append(filteredFolders, folder)
		}
	}

	if len(filteredFolders) == 0 {
		return nil, fmt.Errorf("no such folders found with filter %v", celCmd)
	}

	return filteredFolders, nil
}
