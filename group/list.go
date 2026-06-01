package group

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
	"github.com/spf13/cobra"

	"github.com/pterm/pterm"
)

// GroupListCmd Lists a Passbolt Group
var GroupListCmd = &cobra.Command{
	Use:     "group",
	Short:   "Lists Passbolt Groups",
	Long:    `Lists Passbolt Groups`,
	Aliases: []string{"groups"},
	RunE:    GroupList,
}

func init() {
	flags := GroupListCmd.Flags()
	flags.StringArrayP("user", "u", []string{}, "Groups that are shared with group")
	flags.StringArrayP("manager", "m", []string{}, "Groups that are in folder")
	flags.StringArrayP("column", "c", groupDefaultTableColumns, "Columns to return (default list only for table format; JSON format includes all fields by default).\nPossible Columns: "+strings.Join(groupColumnResolver.Canonical(), ", ")+"\nLegacy PascalCase column names (ID, Name, CreatedTimestamp, ...) remain accepted for backwards compatibility.")
}

type groupListConfig struct {
	users          []string
	managers       []string
	columns        []string
	columnsChanged bool
	jsonOutput     bool
	celFilter      string
}

func GroupList(cmd *cobra.Command, args []string) error {
	config, err := parseGroupListFlags(cmd)
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

	groups, err := client.GetGroups(ctx, &api.GetGroupsOptions{
		FilterHasUsers:    config.users,
		FilterHasManagers: config.managers,
	})
	if err != nil {
		return fmt.Errorf("listing Group: %w", err)
	}

	groups, err = filterGroups(&groups, config.celFilter, ctx)
	if err != nil {
		return err
	}

	if config.jsonOutput {
		return printJSONGroups(groups, config.columnsChanged, config.columns)
	}

	return printTableGroups(config.columns, groups)
}

func printJSONGroups(groups []api.Group, isColumnsChanged bool, columns []string) error {
	outputGroups := make([]GroupJSONOutput, len(groups))
	for i := range groups {
		outputGroups[i] = GroupJSONOutput{
			ID:                &groups[i].ID,
			Name:              &groups[i].Name,
			CreatedTimestamp:  &groups[i].Created.Time,
			ModifiedTimestamp: &groups[i].Modified.Time,
		}
	}

	if isColumnsChanged {
		filteredMap := make([]map[string]interface{}, len(outputGroups))
		for i := range outputGroups {
			filteredMap[i] = make(map[string]interface{})
			data, _ := json.Marshal(outputGroups[i])
			var groupMap map[string]interface{}
			if err := json.Unmarshal(data, &groupMap); err != nil {
				return fmt.Errorf("unmarshaling group: %w", err)
			}

			for _, col := range columns {
				if val, ok := groupMap[col]; ok {
					filteredMap[i][col] = val
				}
			}
		}

		jsonGroups, err := json.MarshalIndent(filteredMap, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonGroups))
		return nil
	}

	jsonGroups, err := json.MarshalIndent(outputGroups, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonGroups))
	return nil
}

func printTableGroups(columns []string, groups []api.Group) error {
	data := pterm.TableData{columns}

	for _, group := range groups {
		entry := make([]string, len(columns))
		for i, col := range columns {
			// Input is normalized by parseGroupListFlags; a miss here is a
			// defensive guard against a future caller that skips that step.
			spec, ok := groupColumnsByName[col]
			if !ok {
				return fmt.Errorf("unknown column: %q", col)
			}
			entry[i] = spec.tableValue(group)
		}
		data = append(data, entry)
	}

	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	return nil
}

func parseGroupListFlags(cmd *cobra.Command) (*groupListConfig, error) {
	users, err := cmd.Flags().GetStringArray("user")
	if err != nil {
		return nil, err
	}
	managers, err := cmd.Flags().GetStringArray("manager")
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
	columns, err = groupColumnResolver.NormalizeAll(columns)
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

	return &groupListConfig{
		users:          users,
		managers:       managers,
		columns:        columns,
		columnsChanged: cmd.Flags().Changed("column"),
		jsonOutput:     jsonOutput,
		celFilter:      celFilter,
	}, nil
}
