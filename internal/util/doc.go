// Package util provides shared utilities for the CLI.
//
// # Shared command lifecycle
//
// WithClient(cmd, fn) wraps context creation, login, session-key persistence +
// logout, and SilenceUsage so command RunE functions don't repeat that plumbing.
// Read flags first (so flag errors still print usage), then run the API calls and
// rendering inside fn, which receives the context and a logged-in *api.Client.
//
// # Shared output helpers
//
// Entity commands render results as JSON or pterm tables. These helpers keep that
// output consistent and remove the per-entity duplication; the typed JSON output
// structs (one per entity) stay in their packages and are passed in generically:
//
//   - PrintJSON / FprintJSON marshal a value as indented JSON (single object or
//     array), to stdout or an explicit io.Writer.
//   - PrintJSONColumnFiltered keeps only the requested keys when the user passed an
//     explicit --column subset.
//   - PrintTable renders a pterm table from a column list and a per-cell value
//     function; it is generic over the item type.
//   - PermissionsToJSONOutput / PrintPermissionTable render the permission output
//     shared by the resource and folder commands.
package util
