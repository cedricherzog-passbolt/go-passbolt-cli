package util

import (
	"errors"
	"fmt"
	"testing"
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
