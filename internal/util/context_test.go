package util

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

// GetContext must derive the request deadline from the configured timeout.
func TestGetContext_UsesConfiguredTimeout(t *testing.T) {
	viper.Reset()
	viper.Set("timeout", 30*time.Second)

	start := time.Now()
	ctx, cancel := GetContext()
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have a deadline")
	}
	got := deadline.Sub(start)
	// Allow a generous tolerance for scheduling jitter between Now() and WithTimeout.
	if got < 29*time.Second || got > 31*time.Second {
		t.Errorf("deadline ~%v from start, want ~30s", got)
	}
}
