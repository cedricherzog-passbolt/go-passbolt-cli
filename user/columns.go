package user

import (
	"strconv"
	"time"

	"al.essio.dev/pkg/shellescape"
	"github.com/google/cel-go/cel"
	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
)

type columnSpec struct {
	name         string
	aliases      []string
	defaultTable bool
	celType      *cel.Type
	celValue     func(u api.User) any
	tableValue   func(u api.User) string
}

var userColumns = []columnSpec{
	{
		name:         "id",
		aliases:      []string{"ID"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(u api.User) any { return u.ID },
		tableValue:   func(u api.User) string { return u.ID },
	},
	{
		name:         "username",
		aliases:      []string{"Username"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(u api.User) any { return u.Username },
		tableValue:   func(u api.User) string { return shellescape.StripUnsafe(u.Username) },
	},
	{
		name:         "first_name",
		aliases:      []string{"FirstName", "firstname"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(u api.User) any { return u.Profile.FirstName },
		tableValue:   func(u api.User) string { return shellescape.StripUnsafe(u.Profile.FirstName) },
	},
	{
		name:         "last_name",
		aliases:      []string{"LastName", "lastname"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(u api.User) any { return u.Profile.LastName },
		tableValue:   func(u api.User) string { return shellescape.StripUnsafe(u.Profile.LastName) },
	},
	{
		name:         "role",
		aliases:      []string{"Role"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(u api.User) any { return u.Role.Name },
		tableValue:   func(u api.User) string { return shellescape.StripUnsafe(u.Role.Name) },
	},
	{
		name:         "created_timestamp",
		aliases:      []string{"CreatedTimestamp", "createdtimestamp"},
		defaultTable: false,
		celType:      cel.TimestampType,
		celValue:     func(u api.User) any { return u.Created.Time },
		tableValue:   func(u api.User) string { return u.Created.Format(time.RFC3339) },
	},
	{
		name:         "modified_timestamp",
		aliases:      []string{"ModifiedTimestamp", "modifiedtimestamp"},
		defaultTable: false,
		celType:      cel.TimestampType,
		celValue:     func(u api.User) any { return u.Modified.Time },
		tableValue:   func(u api.User) string { return u.Modified.Format(time.RFC3339) },
	},
	{
		name:         "active",
		aliases:      []string{"Active"},
		defaultTable: false,
		celType:      cel.BoolType,
		celValue:     func(u api.User) any { return u.Active },
		tableValue:   func(u api.User) string { return strconv.FormatBool(u.Active) },
	},
	{
		name:         "deleted",
		aliases:      []string{"Deleted"},
		defaultTable: false,
		celType:      cel.BoolType,
		celValue:     func(u api.User) any { return u.Deleted },
		tableValue:   func(u api.User) string { return strconv.FormatBool(u.Deleted) },
	},
	{
		// Derived from the nullable *Time field: User.Disabled non-nil ⇒
		// the user is currently disabled. Bool semantics match how users
		// reason about the flag ("is this user disabled?"); an audit-friendly
		// disabled_at timestamp can be added later if needed.
		name:         "disabled",
		aliases:      []string{"Disabled"},
		defaultTable: false,
		celType:      cel.BoolType,
		celValue:     func(u api.User) any { return u.Disabled != nil },
		tableValue:   func(u api.User) string { return strconv.FormatBool(u.Disabled != nil) },
	},
}

var (
	userColumnsByName       = buildUserColumnsByName()
	userColumnResolver      = buildUserColumnResolver()
	userDefaultTableColumns = buildUserDefaultTableColumns()
	userCelEnvOptions       = buildUserCelEnvOptions()
)

func buildUserColumnsByName() map[string]columnSpec {
	m := make(map[string]columnSpec, len(userColumns))
	for _, c := range userColumns {
		m[c.name] = c
	}
	return m
}

func buildUserColumnResolver() *util.ColumnAliasResolver {
	aliases := make([]util.ColumnAlias, len(userColumns))
	for i, c := range userColumns {
		aliases[i] = util.ColumnAlias{Name: c.name, Aliases: c.aliases}
	}
	return util.NewColumnAliasResolver(aliases)
}

func buildUserDefaultTableColumns() []string {
	out := []string{}
	for _, c := range userColumns {
		if c.defaultTable {
			out = append(out, c.name)
		}
	}
	return out
}

func buildUserCelEnvOptions() []cel.EnvOption {
	out := make([]cel.EnvOption, 0, len(userColumns)*2)
	for _, c := range userColumns {
		out = append(out, cel.Variable(c.name, c.celType))
		for _, a := range c.aliases {
			out = append(out, cel.Variable(a, c.celType))
		}
	}
	return out
}

func userCelEvalMap(u api.User) map[string]any {
	m := make(map[string]any, len(userColumns)*2)
	for _, c := range userColumns {
		v := c.celValue(u)
		m[c.name] = v
		for _, a := range c.aliases {
			m[a] = v
		}
	}
	return m
}
