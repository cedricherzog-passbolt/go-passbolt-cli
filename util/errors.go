package util

import "errors"

// Sentinel errors for recurring CLI input-validation failures. They are
// returned (optionally wrapped with extra context via fmt.Errorf and %w) so
// callers and tests can match them with errors.Is rather than comparing
// message strings.
var (
	// ErrNoID is returned by commands that require an --id flag when none is provided.
	ErrNoID = errors.New("no ID provided")

	// ErrNoColumns is returned by list/get commands when no output column is selected.
	ErrNoColumns = errors.New("no columns specified")
)
