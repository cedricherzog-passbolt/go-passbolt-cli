package resource

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/passbolt/go-passbolt-cli/util"
	"github.com/passbolt/go-passbolt/api"
	"github.com/passbolt/go-passbolt/helper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// decryptedResource holds the result of decrypting a single resource
type decryptedResource struct {
	index          int
	resource       api.Resource
	name           string
	username       string
	uri            string
	password       string
	description    string
	metadataFields map[string]any
	secretFields   map[string]any
	err            error
}

// ResourceListCmd Lists a Passbolt Resource
var ResourceListCmd = &cobra.Command{
	Use:     "resource",
	Short:   "Lists Passbolt Resources",
	Long:    `Lists Passbolt Resources`,
	Aliases: []string{"resources"},
	RunE:    ResourceList,
}

func init() {
	flags := ResourceListCmd.Flags()
	flags.Bool("favorite", false, "Resources that are marked as favorite")
	flags.Bool("own", false, "Resources that are owned by me")
	flags.StringP("group", "g", "", "Resources that are shared with group")
	flags.StringArrayP("folder", "f", []string{}, "Resources that are in folder")
	flags.StringArrayP("column", "c", resourceColumns.DefaultTableColumns(), "Columns to return (default list only for table format; JSON format includes all fields by default).\nPossible Columns: "+strings.Join(resourceColumns.Resolver().Canonical(), ", ")+"\nLegacy PascalCase column names (ID, FolderParentID, ...) remain accepted for backwards compatibility.")
}

type resourceListConfig struct {
	favorite       bool
	own            bool
	group          string
	folderParents  []string
	columns        []string
	columnsChanged bool
	jsonOutput     bool
	celFilter      string
}

func ResourceList(cmd *cobra.Command, args []string) error {
	config, err := parseResourceListFlags(cmd)
	if err != nil {
		return err
	}

	// Check if we need to fetch secrets (expensive server join + RSA decryption)
	// For v5 resources, metadata (name, username, uri) can be decrypted without secrets
	needSecrets := resourceColumns.RequiresSecrets(config.columns)

	// Check if CEL filter references any secret-bearing column (canonical or alias).
	if !needSecrets && config.celFilter != "" {
		refsSecrets, err := util.CELExpressionReferencesFields(config.celFilter, resourceColumns.SecretCelNames(), resourceColumns.CelEnvOptions()...)
		if err != nil {
			return fmt.Errorf("parsing filter: %w", err)
		}
		needSecrets = refsSecrets
	}

	return util.WithClient(cmd, func(ctx context.Context, client *api.Client) error {
		resources, err := client.GetResources(ctx, &api.GetResourcesOptions{
			FilterIsFavorite:        config.favorite,
			FilterIsOwnedByMe:       config.own,
			FilterIsSharedWithGroup: config.group,
			FilterHasParent:         config.folderParents,
			ContainSecret:           needSecrets,
		})
		if err != nil {
			return fmt.Errorf("listing Resource: %w", err)
		}

		// Decrypt all resources in parallel
		decrypted, err := decryptResourcesParallel(ctx, client, resources, needSecrets)
		if err != nil {
			return err
		}

		// Apply CEL filter on already-decrypted data
		if config.celFilter != "" {
			decrypted, err = resourceColumns.Filter(ctx, decrypted, config.celFilter)
			if err != nil {
				return err
			}
		}

		if config.jsonOutput {
			return printJSONResources(decrypted, config.columnsChanged, config.columns)
		}

		return printTableResources(decrypted, config.columns)
	})
}

func decryptResourcesParallel(ctx context.Context, client *api.Client, resources []api.Resource, needSecrets bool) ([]decryptedResource, error) {
	// Use parallel decryption with worker pool
	numWorkers := int(viper.GetUint("workers"))

	// Limit Worker count to Resource count
	if len(resources) < numWorkers {
		numWorkers = len(resources)
	}

	// Filter resources - only require secrets if we're fetching them
	var validResources []api.Resource
	for i := range resources {
		if needSecrets && len(resources[i].Secrets) == 0 {
			continue
		}
		validResources = append(validResources, resources[i])
	}

	if len(validResources) == 0 {
		return []decryptedResource{}, nil
	}

	// Channel for work items and results
	// Note: Session keys are pre-fetched during Login() when the server supports v5 metadata,
	// so no additional prefetching is needed here.
	jobs := make(chan int, len(validResources))
	results := make(chan decryptedResource, len(validResources))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				resource := validResources[idx]

				// Lookup resource type from cache (single API call for all types)
				rType, err := client.GetResourceTypeCached(ctx, resource.ResourceTypeID)
				if err != nil {
					results <- decryptedResource{index: idx, err: fmt.Errorf("get ResourceType: %w", err)}
					continue
				}

				// For v4 resources without secret decryption, use plaintext fields directly
				// This avoids unnecessary function calls for 10k+ resources
				isV5 := strings.HasPrefix(rType.Slug, "v5-")
				if !needSecrets && !isV5 {
					// V4 resource - metadata is plaintext, no decryption needed
					results <- decryptedResource{
						index:       idx,
						resource:    resource,
						name:        resource.Name,
						username:    resource.Username,
						uri:         resource.URI,
						password:    "",
						description: resource.Description,
						metadataFields: map[string]any{
							"name":        resource.Name,
							"username":    resource.Username,
							"uri":         resource.URI,
							"description": resource.Description,
						},
					}
					continue
				}

				// Handle case where secrets weren't fetched
				var secret api.Secret
				if len(resource.Secrets) > 0 {
					secret = resource.Secrets[0]
				}

				_, metaFields, secFields, err := helper.GetResourceFieldMaps(
					client,
					resource,
					secret,
					*rType,
					needSecrets,
				)
				results <- decryptedResource{
					index:          idx,
					resource:       resource,
					name:           helper.GetStringField(metaFields, "name"),
					username:       helper.GetStringField(metaFields, "username"),
					uri:            helper.GetStringField(metaFields, "uri"),
					password:       helper.GetStringField(secFields, "password"),
					description:    helper.GetStringField(metaFields, "description"),
					metadataFields: metaFields,
					secretFields:   secFields,
					err:            err,
				}
			}
		}()
	}

	// Send jobs
	for i := range validResources {
		jobs <- i
	}
	close(jobs)

	// Wait for workers and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results first
	allResults := make([]decryptedResource, len(validResources))
	for result := range results {
		allResults[result.index] = result
	}

	// Process results, skipping unsupported types
	decrypted := make([]decryptedResource, 0, len(validResources))
	skippedTypes := make(map[string]int)

	for _, result := range allResults {
		if result.err != nil {
			if errors.Is(result.err, helper.ErrUnsupportedResourceType) {
				// Get type slug for warning message
				rType, _ := client.GetResourceTypeCached(ctx, result.resource.ResourceTypeID)
				typeSlug := "unknown"
				if rType != nil {
					typeSlug = rType.Slug
				}
				skippedTypes[typeSlug]++
				continue
			}
			// Other errors are still fatal
			return nil, fmt.Errorf("get Resource %w", result.err)
		}
		decrypted = append(decrypted, result)
	}

	// Print warning summary to stderr
	if len(skippedTypes) > 0 {
		total := 0
		for _, count := range skippedTypes {
			total += count
		}
		fmt.Fprintf(os.Stderr, "Warning: %d resource(s) skipped due to unsupported types:\n", total)
		for typeSlug, count := range skippedTypes {
			fmt.Fprintf(os.Stderr, "  - %s: %d\n", typeSlug, count)
		}
	}

	return decrypted, nil
}

func printJSONResources(
	decrypted []decryptedResource,
	isColumnsChanged bool,
	columns []string,
) error {
	outputResources := make([]ResourceJSONOutput, len(decrypted))
	for i, d := range decrypted {
		name := d.name
		username := d.username
		uri := d.uri
		pass := d.password
		desc := d.description
		output := ResourceJSONOutput{
			ID:                &d.resource.ID,
			FolderParentID:    &d.resource.FolderParentID,
			Name:              &name,
			Username:          &username,
			URI:               &uri,
			Password:          &pass,
			Description:       &desc,
			CreatedTimestamp:  &d.resource.Created.Time,
			ModifiedTimestamp: &d.resource.Modified.Time,
			Deleted:           d.resource.Deleted,
			Expired:           d.resource.Expired != nil,
			ResourceTypeID:    &d.resource.ResourceTypeID,
		}
		if d.resource.Expired != nil {
			output.ExpiredAt = &d.resource.Expired.Time
		}
		if len(d.metadataFields) > 0 {
			output.Metadata = d.metadataFields
		}
		if len(d.secretFields) > 0 {
			output.Secret = d.secretFields
		}
		outputResources[i] = output
	}

	if isColumnsChanged {
		return util.PrintJSONColumnFiltered(outputResources, columns)
	}

	return util.PrintJSON(outputResources)
}

func printTableResources(
	decrypted []decryptedResource,
	columns []string,
) error {
	// Input is normalized by parseResourceListFlags; a miss in the resolver is a
	// defensive guard against a future caller that skips that step.
	return util.PrintTable(columns, decrypted, resourceColumns.TableValue)
}

func parseResourceListFlags(cmd *cobra.Command) (*resourceListConfig, error) {
	favorite, err := cmd.Flags().GetBool("favorite")
	if err != nil {
		return nil, err
	}
	own, err := cmd.Flags().GetBool("own")
	if err != nil {
		return nil, err
	}
	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}
	folderParents, err := cmd.Flags().GetStringArray("folder")
	if err != nil {
		return nil, err
	}
	columns, err := cmd.Flags().GetStringArray("column")
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, util.NoColumnsError(resourceColumns.Resolver().Canonical())
	}
	columns, err = resourceColumns.Resolver().NormalizeAll(columns)
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

	return &resourceListConfig{
		favorite:       favorite,
		own:            own,
		group:          group,
		folderParents:  folderParents,
		columns:        columns,
		columnsChanged: cmd.Flags().Changed("column"),
		jsonOutput:     jsonOutput,
		celFilter:      celFilter,
	}, nil
}
