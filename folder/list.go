package folder

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
	"github.com/spf13/cobra"

	"github.com/pterm/pterm"
)

// FolderListCmd Lists a Passbolt Folder
var FolderListCmd = &cobra.Command{
	Use:     "folder",
	Short:   "Lists Passbolt Folders",
	Long:    `Lists Passbolt Folders`,
	Aliases: []string{"folders"},
	RunE:    FolderList,
}

func init() {
	flags := FolderListCmd.Flags()
	flags.StringP("search", "s", "", "Folders that have this in the Name")
	flags.StringArrayP("folder", "f", []string{}, "Folders that are in this Folder")
	flags.StringArrayP("group", "g", []string{}, "Folders that are shared with group")
	flags.StringArrayP("column", "c", folderDefaultTableColumns, "Columns to return (default list only for table format; JSON format includes all fields by default).\nPossible Columns: "+strings.Join(folderColumnResolver.Canonical(), ", ")+"\nLegacy PascalCase column names (ID, FolderParentID, ...) remain accepted for backwards compatibility.")
}

type folderListConfig struct {
	search         string
	parentFolders  []string
	columns        []string
	columnsChanged bool
	jsonOutput     bool
	celFilter      string
}

func FolderList(cmd *cobra.Command, args []string) error {
	config, err := parseFolderListFlags(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := util.GetContext()
	defer cancel()

	client, err := util.GetClient(ctx)
	if err != nil {
		return err
	}
	defer util.SaveSessionKeysAndLogout(ctx, client)
	cmd.SilenceUsage = true

	folders, err := client.GetFolders(ctx, &api.GetFoldersOptions{
		FilterHasParent: config.parentFolders,
		FilterSearch:    config.search,
	})
	if err != nil {
		return fmt.Errorf("listing Folder: %w", err)
	}

	folders, err = filterFolders(&folders, config.celFilter, ctx)
	if err != nil {
		return err
	}

	if config.jsonOutput {
		return printJSONFolders(folders, config.columnsChanged, config.columns)
	}

	return printTableFolders(config.columns, folders)
}

func printJSONFolders(folders []api.Folder, isColumnsChanged bool, columns []string) error {
	outputFolders := make([]FolderJSONOutput, len(folders))
	for i := range folders {
		outputFolders[i] = FolderJSONOutput{
			ID:                &folders[i].ID,
			FolderParentID:    &folders[i].FolderParentID,
			Name:              &folders[i].Name,
			CreatedTimestamp:  &folders[i].Created.Time,
			ModifiedTimestamp: &folders[i].Modified.Time,
			Personal:          folders[i].Personal,
		}
	}

	if isColumnsChanged {
		filteredMap := make([]map[string]interface{}, len(outputFolders))
		for i := range outputFolders {
			filteredMap[i] = make(map[string]interface{})
			data, _ := json.Marshal(outputFolders[i])
			var folderMap map[string]interface{}
			if err := json.Unmarshal(data, &folderMap); err != nil {
				return fmt.Errorf("unmarshaling folder: %w", err)
			}

			for _, col := range columns {
				if val, ok := folderMap[col]; ok {
					filteredMap[i][col] = val
				}
			}
		}

		jsonFolders, err := json.MarshalIndent(filteredMap, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonFolders))
		return nil
	}

	jsonFolders, err := json.MarshalIndent(outputFolders, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonFolders))
	return nil
}

func printTableFolders(columns []string, folders []api.Folder) error {
	data := pterm.TableData{columns}

	for _, folder := range folders {
		entry := make([]string, len(columns))
		for i, col := range columns {
			// Input is normalized by parseFolderListFlags; a miss here is a
			// defensive guard against a future caller that skips that step.
			spec, ok := folderColumnsByName[col]
			if !ok {
				return fmt.Errorf("unknown column: %q", col)
			}
			entry[i] = spec.tableValue(folder)
		}
		data = append(data, entry)
	}

	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	return nil
}

func parseFolderListFlags(cmd *cobra.Command) (*folderListConfig, error) {
	search, err := cmd.Flags().GetString("search")
	if err != nil {
		return nil, err
	}
	parentFolders, err := cmd.Flags().GetStringArray("folder")
	if err != nil {
		return nil, err
	}
	columns, err := cmd.Flags().GetStringArray("column")
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("you need to specify at least one column to return")
	}
	columns, err = folderColumnResolver.NormalizeAll(columns)
	if err != nil {
		return nil, err
	}
	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return nil, err
	}
	celFilter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return nil, err
	}

	return &folderListConfig{
		search:         search,
		parentFolders:  parentFolders,
		columns:        columns,
		columnsChanged: cmd.Flags().Changed("column"),
		jsonOutput:     jsonOutput,
		celFilter:      celFilter,
	}, nil
}
