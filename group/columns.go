package group

import (
	"time"

	"al.essio.dev/pkg/shellescape"
	"github.com/google/cel-go/cel"
	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
)

// columnSpec describes a single --column / --filter attribute on Group.
// Canonical name is snake_case (matches JSON tag and Passbolt API
// conventions); aliases keep legacy PascalCase and lowercased-concat forms
// working for back-compat.
type columnSpec struct {
	name         string
	aliases      []string
	defaultTable bool
	celType      *cel.Type
	celValue     func(g api.Group) any
	tableValue   func(g api.Group) string
}

var groupColumns = []columnSpec{
	{
		name:         "id",
		aliases:      []string{"ID"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(g api.Group) any { return g.ID },
		tableValue:   func(g api.Group) string { return g.ID },
	},
	{
		name:         "name",
		aliases:      []string{"Name"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(g api.Group) any { return g.Name },
		tableValue:   func(g api.Group) string { return shellescape.StripUnsafe(g.Name) },
	},
	{
		name:         "created_timestamp",
		aliases:      []string{"CreatedTimestamp", "createdtimestamp"},
		defaultTable: false,
		celType:      cel.TimestampType,
		celValue:     func(g api.Group) any { return g.Created.Time },
		tableValue:   func(g api.Group) string { return g.Created.Format(time.RFC3339) },
	},
	{
		name:         "modified_timestamp",
		aliases:      []string{"ModifiedTimestamp", "modifiedtimestamp"},
		defaultTable: false,
		celType:      cel.TimestampType,
		celValue:     func(g api.Group) any { return g.Modified.Time },
		tableValue:   func(g api.Group) string { return g.Modified.Format(time.RFC3339) },
	},
}

var (
	groupColumnsByName       = buildGroupColumnsByName()
	groupColumnResolver      = buildGroupColumnResolver()
	groupDefaultTableColumns = buildGroupDefaultTableColumns()
	groupCelEnvOptions       = buildGroupCelEnvOptions()
)

func buildGroupColumnsByName() map[string]columnSpec {
	m := make(map[string]columnSpec, len(groupColumns))
	for _, c := range groupColumns {
		m[c.name] = c
	}
	return m
}

func buildGroupColumnResolver() *util.ColumnAliasResolver {
	aliases := make([]util.ColumnAlias, len(groupColumns))
	for i, c := range groupColumns {
		aliases[i] = util.ColumnAlias{Name: c.name, Aliases: c.aliases}
	}
	return util.NewColumnAliasResolver(aliases)
}

func buildGroupDefaultTableColumns() []string {
	out := []string{}
	for _, c := range groupColumns {
		if c.defaultTable {
			out = append(out, c.name)
		}
	}
	return out
}

// buildGroupCelEnvOptions registers each canonical column as a CEL variable
// plus each alias, so legacy filter expressions like 'Name == "X"' continue
// to compile alongside canonical 'name == "X"'.
func buildGroupCelEnvOptions() []cel.EnvOption {
	out := make([]cel.EnvOption, 0, len(groupColumns)*2)
	for _, c := range groupColumns {
		out = append(out, cel.Variable(c.name, c.celType))
		for _, a := range c.aliases {
			out = append(out, cel.Variable(a, c.celType))
		}
	}
	return out
}

// groupCelEvalMap builds the variable→value map passed to ContextEval,
// populating both canonical and alias keys with the same value.
func groupCelEvalMap(g api.Group) map[string]any {
	m := make(map[string]any, len(groupColumns)*2)
	for _, c := range groupColumns {
		v := c.celValue(g)
		m[c.name] = v
		for _, a := range c.aliases {
			m[a] = v
		}
	}
	return m
}
