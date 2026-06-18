package util

import (
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// A valid self-signed ECDSA cert/key pair used to exercise the mTLS happy path.
// tls.X509KeyPair does not check expiry, so this fixture is date-independent.
const (
	testClientCertPEM = `-----BEGIN CERTIFICATE-----
MIIBHjCBxaADAgECAgEBMAoGCCqGSM49BAMCMBgxFjAUBgNVBAMTDXBhc3Nib2x0
LXRlc3QwIBcNNzAwMTAxMDAwMDAwWhgPMjI0MjAzMTYxMjU2MzJaMBgxFjAUBgNV
BAMTDXBhc3Nib2x0LXRlc3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARLeZl6
ZkMYW/0t01mI7Ro+8q29K0k3snQ4sDfAu+ep51irAeg5htEDQuzus0xd5Mbi3rSI
g8++0ZsYI2iIXRlPMAoGCCqGSM49BAMCA0gAMEUCIAnUoAgcbCFioE0eSa29kNS9
/jzx1NoFVXaS6tRFlDtxAiEA81bXDibeGJiNBKHXbIuZe1DVHuPnugcDBdT9lrOl
IJA=
-----END CERTIFICATE-----
`
	testClientKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIPNNsFsQUlivmEwap4trO0bOIYRp0AedTvLcVztXCaeDoAoGCCqGSM49
AwEHoUQDQgAES3mZemZDGFv9LdNZiO0aPvKtvStJN7J0OLA3wLvnqedYqwHoOYbR
A0Ls7rNMXeTG4t60iIPPvtGbGCNoiF0ZTw==
-----END EC PRIVATE KEY-----
`
)

func TestGetClientCertificate(t *testing.T) {
	cases := []struct {
		name      string
		cert      string
		key       string
		wantErr   string // substring; "" means no error
		wantCerts int    // number of certs in the returned tls.Certificate
	}{
		{name: "both empty", cert: "", key: "", wantErr: "", wantCerts: 0},
		{name: "cert without key", cert: testClientCertPEM, key: "", wantErr: "private key is empty"},
		{name: "key without cert", cert: "", key: testClientKeyPEM, wantErr: "cert is empty"},
		{name: "invalid pem pair", cert: "not-a-cert", key: "not-a-key", wantErr: "failed to find"},
		{name: "valid pair", cert: testClientCertPEM, key: testClientKeyPEM, wantErr: "", wantCerts: 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("tlsClientCert", c.cert)
			viper.Set("tlsClientPrivateKey", c.key)

			got, err := GetClientCertificate()
			if c.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", c.wantErr)
				}
				if !strings.Contains(err.Error(), c.wantErr) {
					t.Errorf("error %q should contain %q", err.Error(), c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.Certificate) != c.wantCerts {
				t.Errorf("got %d certs, want %d", len(got.Certificate), c.wantCerts)
			}
		})
	}
}

func TestGetHTTPClient_SkipVerify(t *testing.T) {
	for _, skip := range []bool{true, false} {
		viper.Reset()
		viper.Set("tlsSkipVerify", skip)

		client, err := GetHTTPClient()
		if err != nil {
			t.Fatalf("skip=%v: unexpected error: %v", skip, err)
		}
		tr, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("skip=%v: transport is %T, want *http.Transport", skip, client.Transport)
		}
		if tr.TLSClientConfig.InsecureSkipVerify != skip {
			t.Errorf("InsecureSkipVerify = %v, want %v", tr.TLSClientConfig.InsecureSkipVerify, skip)
		}
	}
}

func TestGetHTTPClient_PropagatesCertError(t *testing.T) {
	viper.Reset()
	// Cert without key is an invalid mTLS configuration; GetHTTPClient must surface it.
	viper.Set("tlsClientCert", testClientCertPEM)

	if _, err := GetHTTPClient(); err == nil {
		t.Fatal("expected error from invalid client certificate configuration")
	}
}

func TestGetHTTPClient_EmbedsClientCert(t *testing.T) {
	viper.Reset()
	viper.Set("tlsClientCert", testClientCertPEM)
	viper.Set("tlsClientPrivateKey", testClientKeyPEM)

	client, err := GetHTTPClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tr := client.Transport.(*http.Transport)
	if len(tr.TLSClientConfig.Certificates) != 1 {
		t.Errorf("TLSClientConfig.Certificates = %d, want 1", len(tr.TLSClientConfig.Certificates))
	}
}
