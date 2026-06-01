package resource

import (
	"context"
	"fmt"

	"github.com/passbolt/go-passbolt-cli/util"
)

// filterDecryptedResources filters already-decrypted resources by evaluating a CEL expression.
func filterDecryptedResources(resources []decryptedResource, celCmd string, ctx context.Context) ([]decryptedResource, error) {
	if celCmd == "" {
		return resources, nil
	}

	program, err := util.InitCELProgram(celCmd, resourceCelEnvOptions...)
	if err != nil {
		return nil, err
	}

	filtered := []decryptedResource{}
	for _, d := range resources {
		val, _, err := (*program).ContextEval(ctx, resourceCelEvalMap(d))
		if err != nil {
			return nil, err
		}

		if val.Value() == true {
			filtered = append(filtered, d)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no such resources found with filter %v", celCmd)
	}
	return filtered, nil
}
