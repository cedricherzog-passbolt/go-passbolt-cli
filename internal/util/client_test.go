package util

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// GetClient validates required configuration before any password prompt or
// network call, so these error paths are hermetic.
func TestGetClient_RequiredFieldValidation(t *testing.T) {
	cases := []struct {
		name           string
		serverAddress  string
		userPrivateKey string
		userPassword   string
		wantErr        string
	}{
		{
			name:    "missing serverAddress",
			wantErr: "serverAddress is not defined",
		},
		{
			name:          "missing userPrivateKey",
			serverAddress: "https://passbolt.example.com",
			wantErr:       "userPrivateKey is not defined",
		},
		{
			name:           "invalid userPrivateKey",
			serverAddress:  "https://passbolt.example.com",
			userPrivateKey: "not-a-real-pgp-key",
			userPassword:   "passphrase", // set so GetClient does not prompt on stdin
			wantErr:        "creating Client",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("serverAddress", c.serverAddress)
			viper.Set("userPrivateKey", c.userPrivateKey)
			viper.Set("userPassword", c.userPassword)

			_, err := GetClient(context.Background())
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("error %q should contain %q", err.Error(), c.wantErr)
			}
		})
	}
}
