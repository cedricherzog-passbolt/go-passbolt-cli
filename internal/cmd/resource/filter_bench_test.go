package resource

// Benchmarks the resource-specific column registry + CEL filter end-to-end over
// a synthetic decryptedResource list. This is the real `resource list --filter`
// hot path: resourceColumns.Filter compiles the expression once then evaluates
// every decrypted resource's columns per item. It also backs the
// PROFILE_CLI_PKG=./internal/cmd/resource/ profiling target.

import (
	"context"
	"fmt"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// benchResourceSizes mirrors the util-package tiers: small / large / enterprise
// vault sizes so the per-item eval cost is visible as the list grows.
var benchResourceSizes = []int{100, 1000, 10000}

// benchResourceSink defeats dead-code elimination for the filtered result.
var benchResourceSink int

// benchResources builds n decrypted resources. Created/Modified must be set:
// the timestamp columns deref them during CelValue evaluation (see the note in
// filter_test.go).
func benchResources(n int) []decryptedResource {
	items := make([]decryptedResource, n)
	for i := range items {
		items[i] = decryptedResource{
			resource: api.Resource{
				ID:       fmt.Sprintf("id-%d", i),
				Created:  &api.Time{},
				Modified: &api.Time{},
				Deleted:  i%2 == 0,
			},
			name:        fmt.Sprintf("resource-%d", i),
			username:    fmt.Sprintf("user-%d@example.com", i),
			uri:         fmt.Sprintf("https://host-%d.example.com", i),
			password:    fmt.Sprintf("p%d", i),
			description: "synthetic benchmark resource",
		}
	}
	return items
}

// BenchmarkResourceColumns_Filter measures the resource list filter across vault
// sizes. The expression matches roughly half the resources so Filter never hits
// its zero-match error path.
func BenchmarkResourceColumns_Filter(b *testing.B) {
	ctx := context.Background()
	const expr = `username.contains("@example.com") && !deleted`

	for _, n := range benchResourceSizes {
		items := benchResources(n)
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				out, err := resourceColumns.Filter(ctx, items, expr)
				if err != nil {
					b.Fatalf("Filter: %v", err)
				}
				benchResourceSink = len(out)
			}
		})
	}
}
