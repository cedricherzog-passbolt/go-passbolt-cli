package user

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
	"github.com/spf13/cobra"

	"github.com/pterm/pterm"
)

// UserListCmd Lists a Passbolt User
var UserListCmd = &cobra.Command{
	Use:     "user",
	Short:   "Lists Passbolt Users",
	Long:    `Lists Passbolt Users`,
	Aliases: []string{"users"},
	RunE:    UserList,
}

func init() {
	flags := UserListCmd.Flags()
	flags.StringArrayP("group", "g", []string{}, "Users that are members of groups")
	flags.StringArrayP("resource", "r", []string{}, "Users that have access to resources")
	flags.StringP("search", "s", "", "Search for Users")
	flags.BoolP("admin", "a", false, "Only show Admins")
	flags.StringArrayP("column", "c", userDefaultTableColumns, "Columns to return (default list only for table format; JSON format includes all fields by default).\nPossible Columns: "+strings.Join(userColumnResolver.Canonical(), ", ")+"\nLegacy PascalCase column names (ID, FirstName, ...) remain accepted for backwards compatibility.")
}

type userListConfig struct {
	groups         []string
	resources      []string
	search         string
	admin          bool
	columns        []string
	columnsChanged bool
	jsonOutput     bool
	celFilter      string
}

func UserList(cmd *cobra.Command, args []string) error {
	config, err := parseUserListFlags(cmd)
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

	users, err := client.GetUsers(ctx, &api.GetUsersOptions{
		FilterHasGroup:  config.groups,
		FilterHasAccess: config.resources,
		FilterSearch:    config.search,
		FilterIsAdmin:   config.admin,
	})
	if err != nil {
		return fmt.Errorf("listing User: %w", err)
	}

	users, err = filterUsers(&users, config.celFilter, ctx)
	if err != nil {
		return err
	}

	if config.jsonOutput {
		return printJSONUsers(users, config.columnsChanged, config.columns)
	}

	return printTableUsers(config.columns, users)
}

func printJSONUsers(users []api.User, isColumnsChanged bool, columns []string) error {
	outputUsers := make([]UserJSONOutput, len(users))
	for i := range users {
		outputUsers[i] = UserJSONOutput{
			ID:                &users[i].ID,
			Username:          &users[i].Username,
			FirstName:         &users[i].Profile.FirstName,
			LastName:          &users[i].Profile.LastName,
			Role:              &users[i].Role.Name,
			CreatedTimestamp:  &users[i].Created.Time,
			ModifiedTimestamp: &users[i].Modified.Time,
			Active:            users[i].Active,
			Deleted:           users[i].Deleted,
			Disabled:          users[i].Disabled != nil,
		}
	}

	if isColumnsChanged {
		filteredMap := make([]map[string]interface{}, len(outputUsers))
		for i := range outputUsers {
			filteredMap[i] = make(map[string]interface{})
			data, _ := json.Marshal(outputUsers[i])
			var userMap map[string]interface{}
			if err := json.Unmarshal(data, &userMap); err != nil {
				return fmt.Errorf("unmarshaling user: %w", err)
			}

			for _, col := range columns {
				if val, ok := userMap[col]; ok {
					filteredMap[i][col] = val
				}
			}
		}

		jsonUsers, err := json.MarshalIndent(filteredMap, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonUsers))
		return nil
	}

	jsonUsers, err := json.MarshalIndent(outputUsers, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonUsers))
	return nil
}

func printTableUsers(columns []string, users []api.User) error {
	data := pterm.TableData{columns}

	for _, user := range users {
		entry := make([]string, len(columns))
		for i, col := range columns {
			// Input is normalized by parseUserListFlags; a miss here is a
			// defensive guard against a future caller that skips that step.
			spec, ok := userColumnsByName[col]
			if !ok {
				return fmt.Errorf("unknown column: %q", col)
			}
			entry[i] = spec.tableValue(user)
		}
		data = append(data, entry)
	}

	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	return nil
}

func parseUserListFlags(cmd *cobra.Command) (*userListConfig, error) {
	groups, err := cmd.Flags().GetStringArray("group")
	if err != nil {
		return nil, err
	}
	resources, err := cmd.Flags().GetStringArray("resource")
	if err != nil {
		return nil, err
	}
	search, err := cmd.Flags().GetString("search")
	if err != nil {
		return nil, err
	}
	admin, err := cmd.Flags().GetBool("admin")
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
	columns, err = userColumnResolver.NormalizeAll(columns)
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

	return &userListConfig{
		groups:         groups,
		resources:      resources,
		search:         search,
		admin:          admin,
		columns:        columns,
		columnsChanged: cmd.Flags().Changed("column"),
		jsonOutput:     jsonOutput,
		celFilter:      celFilter,
	}, nil
}
