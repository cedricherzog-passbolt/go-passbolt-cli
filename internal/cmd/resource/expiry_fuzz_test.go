package resource

import (
	"testing"
	"time"
)

// FuzzParseExpiry fuzzes the dual-mode expiry parser (absolute RFC3339 vs. Go
// duration). Invariant: it never panics, and whenever it returns a value without
// error for non-empty input, that value is a valid RFC3339 timestamp — so callers
// can always hand the result straight to the API.
func FuzzParseExpiry(f *testing.F) {
	f.Add("2021-01-01T00:00:00Z")
	f.Add("2021-01-01T00:00:00.5+02:00")
	f.Add("48h")
	f.Add("30m")
	f.Add("")
	f.Add("not-a-time")
	f.Add("-1h")

	f.Fuzz(func(t *testing.T, in string) {
		got, err := ParseExpiry(in)
		if err != nil {
			return
		}
		if in == "" {
			if got != "" {
				t.Fatalf("empty input produced non-empty result %q", got)
			}
			return
		}
		if _, err := time.Parse(time.RFC3339, got); err != nil {
			t.Fatalf("ParseExpiry(%q) = %q which is not valid RFC3339: %v", in, got, err)
		}
	})
}

// FuzzIsUUID guards the CLI-side UUID check used to gate unsafe URL construction
// in SetResourceExpiry. Invariant: it never panics, and anything it accepts is
// exactly 36 bytes (the canonical 8-4-4-4-12 form).
func FuzzIsUUID(f *testing.F) {
	f.Add("11111111-1111-1111-1111-111111111111")
	f.Add("AABBCCDD-1234-5678-9abc-def012345678")
	f.Add("")
	f.Add("../../etc/passwd")

	f.Fuzz(func(t *testing.T, in string) {
		if isUUID(in) && len(in) != 36 {
			t.Fatalf("isUUID accepted non-canonical-length value (%d bytes): %q", len(in), in)
		}
	})
}
