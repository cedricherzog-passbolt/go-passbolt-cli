package resource

import (
	"encoding/json"
	"strconv"
	"time"

	"al.essio.dev/pkg/shellescape"
	"github.com/google/cel-go/cel"
	"github.com/passbolt/go-passbolt-cli/util"
)

// columnSpec describes a single --column / --filter attribute on Resource.
// requiresSecrets marks columns whose value depends on the (expensive) secret
// decryption path — used to decide whether to fetch secrets up-front.
type columnSpec struct {
	name            string
	aliases         []string
	defaultTable    bool
	requiresSecrets bool
	celType         *cel.Type
	celValue        func(d decryptedResource) any
	tableValue      func(d decryptedResource) string
}

func marshalMapForTable(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

var resourceColumns = []columnSpec{
	{
		name:         "id",
		aliases:      []string{"ID"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(d decryptedResource) any { return d.resource.ID },
		tableValue:   func(d decryptedResource) string { return d.resource.ID },
	},
	{
		name:         "folder_parent_id",
		aliases:      []string{"FolderParentID", "folderparentid"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(d decryptedResource) any { return d.resource.FolderParentID },
		tableValue:   func(d decryptedResource) string { return d.resource.FolderParentID },
	},
	{
		name:         "name",
		aliases:      []string{"Name"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(d decryptedResource) any { return d.name },
		tableValue:   func(d decryptedResource) string { return shellescape.StripUnsafe(d.name) },
	},
	{
		name:         "username",
		aliases:      []string{"Username"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(d decryptedResource) any { return d.username },
		tableValue:   func(d decryptedResource) string { return shellescape.StripUnsafe(d.username) },
	},
	{
		name:         "uri",
		aliases:      []string{"URI"},
		defaultTable: true,
		celType:      cel.StringType,
		celValue:     func(d decryptedResource) any { return d.uri },
		tableValue:   func(d decryptedResource) string { return shellescape.StripUnsafe(d.uri) },
	},
	{
		name:            "password",
		aliases:         []string{"Password"},
		defaultTable:    false,
		requiresSecrets: true,
		celType:         cel.StringType,
		celValue:        func(d decryptedResource) any { return d.password },
		tableValue:      func(d decryptedResource) string { return shellescape.StripUnsafe(d.password) },
	},
	{
		name:            "description",
		aliases:         []string{"Description"},
		defaultTable:    false,
		requiresSecrets: true,
		celType:         cel.StringType,
		celValue:        func(d decryptedResource) any { return d.description },
		tableValue:      func(d decryptedResource) string { return shellescape.StripUnsafe(d.description) },
	},
	{
		name:         "created_timestamp",
		aliases:      []string{"CreatedTimestamp", "createdtimestamp"},
		defaultTable: false,
		celType:      cel.TimestampType,
		celValue:     func(d decryptedResource) any { return d.resource.Created.Time },
		tableValue:   func(d decryptedResource) string { return d.resource.Created.Format(time.RFC3339) },
	},
	{
		name:         "modified_timestamp",
		aliases:      []string{"ModifiedTimestamp", "modifiedtimestamp"},
		defaultTable: false,
		celType:      cel.TimestampType,
		celValue:     func(d decryptedResource) any { return d.resource.Modified.Time },
		tableValue:   func(d decryptedResource) string { return d.resource.Modified.Format(time.RFC3339) },
	},
	{
		name:         "metadata",
		aliases:      []string{"Metadata"},
		defaultTable: false,
		celType:      cel.MapType(cel.StringType, cel.DynType),
		celValue: func(d decryptedResource) any {
			if d.metadataFields == nil {
				return map[string]any{}
			}
			return d.metadataFields
		},
		tableValue: func(d decryptedResource) string { return marshalMapForTable(d.metadataFields) },
	},
	{
		name:            "secret",
		aliases:         []string{"Secret"},
		defaultTable:    false,
		requiresSecrets: true,
		celType:         cel.MapType(cel.StringType, cel.DynType),
		celValue: func(d decryptedResource) any {
			if d.secretFields == nil {
				return map[string]any{}
			}
			return d.secretFields
		},
		tableValue: func(d decryptedResource) string { return marshalMapForTable(d.secretFields) },
	},
	{
		name:         "deleted",
		aliases:      []string{"Deleted"},
		defaultTable: false,
		celType:      cel.BoolType,
		celValue:     func(d decryptedResource) any { return d.resource.Deleted },
		tableValue:   func(d decryptedResource) string { return strconv.FormatBool(d.resource.Deleted) },
	},
	{
		// Derived bool: true iff Resource.Expired is set. Pairs with the
		// expired_at column below which exposes the actual timestamp.
		name:         "expired",
		aliases:      []string{"Expired"},
		defaultTable: false,
		celType:      cel.BoolType,
		celValue:     func(d decryptedResource) any { return d.resource.Expired != nil },
		tableValue:   func(d decryptedResource) string { return strconv.FormatBool(d.resource.Expired != nil) },
	},
	{
		// Nullable timestamp: zero time when not expired. CEL users compare
		// against timestamp() literals; table renders RFC3339 or empty.
		name:         "expired_at",
		aliases:      []string{"ExpiredAt", "expiredat"},
		defaultTable: false,
		celType:      cel.TimestampType,
		celValue: func(d decryptedResource) any {
			if d.resource.Expired == nil {
				return time.Time{}
			}
			return d.resource.Expired.Time
		},
		tableValue: func(d decryptedResource) string {
			if d.resource.Expired == nil {
				return ""
			}
			return d.resource.Expired.Format(time.RFC3339)
		},
	},
	{
		name:         "resource_type_id",
		aliases:      []string{"ResourceTypeID", "resourcetypeid"},
		defaultTable: false,
		celType:      cel.StringType,
		celValue:     func(d decryptedResource) any { return d.resource.ResourceTypeID },
		tableValue:   func(d decryptedResource) string { return d.resource.ResourceTypeID },
	},
}

var (
	resourceColumnsByName       = buildResourceColumnsByName()
	resourceColumnResolver      = buildResourceColumnResolver()
	resourceDefaultTableColumns = buildResourceDefaultTableColumns()
	resourceCelEnvOptions       = buildResourceCelEnvOptions()
	resourceSecretCelNames      = buildResourceSecretCelNames()
)

func buildResourceColumnsByName() map[string]columnSpec {
	m := make(map[string]columnSpec, len(resourceColumns))
	for _, c := range resourceColumns {
		m[c.name] = c
	}
	return m
}

func buildResourceColumnResolver() *util.ColumnAliasResolver {
	aliases := make([]util.ColumnAlias, len(resourceColumns))
	for i, c := range resourceColumns {
		aliases[i] = util.ColumnAlias{Name: c.name, Aliases: c.aliases}
	}
	return util.NewColumnAliasResolver(aliases)
}

func buildResourceDefaultTableColumns() []string {
	out := []string{}
	for _, c := range resourceColumns {
		if c.defaultTable {
			out = append(out, c.name)
		}
	}
	return out
}

func buildResourceCelEnvOptions() []cel.EnvOption {
	out := make([]cel.EnvOption, 0, len(resourceColumns)*2)
	for _, c := range resourceColumns {
		out = append(out, cel.Variable(c.name, c.celType))
		for _, a := range c.aliases {
			out = append(out, cel.Variable(a, c.celType))
		}
	}
	return out
}

// buildResourceSecretCelNames returns all CEL variable names (canonical and
// alias) whose value comes from secret decryption — used to detect whether a
// user's --filter expression forces secret fetching.
func buildResourceSecretCelNames() []string {
	out := []string{}
	for _, c := range resourceColumns {
		if !c.requiresSecrets {
			continue
		}
		out = append(out, c.name)
		out = append(out, c.aliases...)
	}
	return out
}

func resourceCelEvalMap(d decryptedResource) map[string]any {
	m := make(map[string]any, len(resourceColumns)*2)
	for _, c := range resourceColumns {
		v := c.celValue(d)
		m[c.name] = v
		for _, a := range c.aliases {
			m[a] = v
		}
	}
	return m
}

// columnsRequireSecrets reports whether any of the (already-normalized,
// canonical) columns demands secret decryption.
func columnsRequireSecrets(columns []string) bool {
	for _, col := range columns {
		spec, ok := resourceColumnsByName[col]
		if ok && spec.requiresSecrets {
			return true
		}
	}
	return false
}
