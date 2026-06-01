package folder

import (
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
	celValue     func(f api.Folder) any
	tableValue   func(f api.Folder) string
}

var folderColumns = []columnSpec{
	{
		name:         "id",
		aliases:      []string{"ID"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(f api.Folder) any { return f.ID },
		tableValue:   func(f api.Folder) string { return f.ID },
	},
	{
		name:         "folder_parent_id",
		aliases:      []string{"FolderParentID", "folderparentid"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(f api.Folder) any { return f.FolderParentID },
		tableValue:   func(f api.Folder) string { return f.FolderParentID },
	},
	{
		name:         "name",
		aliases:      []string{"Name"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(f api.Folder) any { return f.Name },
		tableValue:   func(f api.Folder) string { return shellescape.StripUnsafe(f.Name) },
	},
	{
		name:         "created_timestamp",
		aliases:      []string{"CreatedTimestamp", "createdtimestamp"},
		defaultTable: false,
		celType:      cel.TimestampType,
		celValue:     func(f api.Folder) any { return f.Created.Time },
		tableValue:   func(f api.Folder) string { return f.Created.Format(time.RFC3339) },
	},
	{
		name:         "modified_timestamp",
		aliases:      []string{"ModifiedTimestamp", "modifiedtimestamp"},
		defaultTable: false,
		celType:      cel.TimestampType,
		celValue:     func(f api.Folder) any { return f.Modified.Time },
		tableValue:   func(f api.Folder) string { return f.Modified.Format(time.RFC3339) },
	},
}

var (
	folderColumnsByName       = buildFolderColumnsByName()
	folderColumnResolver      = buildFolderColumnResolver()
	folderDefaultTableColumns = buildFolderDefaultTableColumns()
	folderCelEnvOptions       = buildFolderCelEnvOptions()
)

func buildFolderColumnsByName() map[string]columnSpec {
	m := make(map[string]columnSpec, len(folderColumns))
	for _, c := range folderColumns {
		m[c.name] = c
	}
	return m
}

func buildFolderColumnResolver() *util.ColumnAliasResolver {
	aliases := make([]util.ColumnAlias, len(folderColumns))
	for i, c := range folderColumns {
		aliases[i] = util.ColumnAlias{Name: c.name, Aliases: c.aliases}
	}
	return util.NewColumnAliasResolver(aliases)
}

func buildFolderDefaultTableColumns() []string {
	out := []string{}
	for _, c := range folderColumns {
		if c.defaultTable {
			out = append(out, c.name)
		}
	}
	return out
}

func buildFolderCelEnvOptions() []cel.EnvOption {
	out := make([]cel.EnvOption, 0, len(folderColumns)*2)
	for _, c := range folderColumns {
		out = append(out, cel.Variable(c.name, c.celType))
		for _, a := range c.aliases {
			out = append(out, cel.Variable(a, c.celType))
		}
	}
	return out
}

func folderCelEvalMap(f api.Folder) map[string]any {
	m := make(map[string]any, len(folderColumns)*2)
	for _, c := range folderColumns {
		v := c.celValue(f)
		m[c.name] = v
		for _, a := range c.aliases {
			m[a] = v
		}
	}
	return m
}
