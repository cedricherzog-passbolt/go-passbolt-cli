package util

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// TestSentinelsMatchableWhenWrapped documents the contract that callers and
// tests rely on: the sentinel errors stay matchable with errors.Is even after
// being wrapped with additional context via fmt.Errorf and %w.
func TestSentinelsMatchableWhenWrapped(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
	}{
		{"ErrNoID", ErrNoID},
		{"ErrNoColumns", ErrNoColumns},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Bare sentinel matches itself.
			if !errors.Is(tt.sentinel, tt.sentinel) {
				t.Fatalf("bare sentinel %v does not match itself", tt.sentinel)
			}
			// Wrapped with %w stays matchable.
			wrapped := fmt.Errorf("doing something: %w", tt.sentinel)
			if !errors.Is(wrapped, tt.sentinel) {
				t.Errorf("wrapped error %q is not matchable with errors.Is", wrapped)
			}
			// Wrapped with extra context (the %w + %v idiom) stays matchable.
			withContext := fmt.Errorf("%w (extra context %v)", tt.sentinel, 42)
			if !errors.Is(withContext, tt.sentinel) {
				t.Errorf("context-wrapped error %q is not matchable with errors.Is", withContext)
			}
		})
	}
}

// TestSentinelsAreDistinct guards against accidentally aliasing the sentinels
// to the same value, which would make errors.Is matching ambiguous.
func TestSentinelsAreDistinct(t *testing.T) {
	if errors.Is(ErrNoID, ErrNoColumns) || errors.Is(ErrNoColumns, ErrNoID) {
		t.Error("ErrNoID and ErrNoColumns must be distinct sentinel errors")
	}
}

// TestNoColumnsError checks the helper lists the valid columns and stays
// matchable with errors.Is(ErrNoColumns).
func TestNoColumnsError(t *testing.T) {
	err := NoColumnsError([]string{"ID", "Name"})
	if !errors.Is(err, ErrNoColumns) {
		t.Errorf("NoColumnsError must wrap ErrNoColumns, got %q", err)
	}
	if !strings.Contains(err.Error(), "ID") || !strings.Contains(err.Error(), "Name") {
		t.Errorf("NoColumnsError should list valid columns, got %q", err)
	}
}

// TestExplainAPIError verifies a friendly hint is added for known status codes
// while the underlying *api.APIError stays recoverable with errors.As, and that
// non-API errors are merely op-wrapped.
func TestExplainAPIError(t *testing.T) {
	apiErr := &api.APIError{StatusCode: 403, Message: "Forbidden"}
	err := ExplainAPIError("logging in", apiErr)
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("expected friendly 403 hint, got %q", err.Error())
	}
	var got *api.APIError
	if !errors.As(err, &got) {
		t.Error("ExplainAPIError must preserve the *api.APIError in the chain")
	}

	base := errors.New("boom")
	wrapped := ExplainAPIError("doing thing", base)
	if !errors.Is(wrapped, base) {
		t.Error("ExplainAPIError must preserve a non-API error chain")
	}
	if !strings.HasPrefix(wrapped.Error(), "doing thing: ") {
		t.Errorf("expected op prefix, got %q", wrapped.Error())
	}
}

// TestAPIStatusHint covers every branch of the status→hint mapping, documenting
// the exact user-facing guidance per code. "" means no hint applies.
func TestAPIStatusHint(t *testing.T) {
	cases := []struct {
		code int
		want string // substring; "" asserts an empty hint
	}{
		{401, "authentication failed"},
		{403, "access denied"},
		{404, "not found"},
		{500, "internal error"},
		{503, "internal error"},
		{400, ""},
		{422, ""},
	}
	for _, c := range cases {
		got := apiStatusHint(c.code)
		if c.want == "" {
			if got != "" {
				t.Errorf("apiStatusHint(%d) = %q, want empty", c.code, got)
			}
			continue
		}
		if !strings.Contains(got, c.want) {
			t.Errorf("apiStatusHint(%d) = %q, want it to contain %q", c.code, got, c.want)
		}
	}
}

// TestExplainAPIError_NoHintStatus verifies that an API error with a status that
// has no hint (400) is only op-wrapped — no spurious hint text — while the
// *api.APIError stays recoverable via errors.As.
func TestExplainAPIError_NoHintStatus(t *testing.T) {
	apiErr := &api.APIError{StatusCode: 400, Message: "Bad Request"}
	err := ExplainAPIError("doing thing", apiErr)

	if !strings.HasPrefix(err.Error(), "doing thing: ") {
		t.Errorf("expected op prefix, got %q", err.Error())
	}
	var got *api.APIError
	if !errors.As(err, &got) {
		t.Error("ExplainAPIError must preserve the *api.APIError in the chain")
	}
	// The message should be just the op wrap around the API error, with no hint
	// phrasing injected.
	if strings.Contains(err.Error(), "authentication failed") ||
		strings.Contains(err.Error(), "access denied") ||
		strings.Contains(err.Error(), "internal error") {
		t.Errorf("no hint should be added for status 400, got %q", err.Error())
	}
}
