package resource

import (
	"strings"
	"testing"
)

// FuzzParseKeyValue fuzzes the "key=value" parser, whose value half is decoded
// as JSON when it looks like a JSON array/object and otherwise kept as a literal
// string. Invariant: it never panics, and on success the input must begin with
// exactly "key=" — the parser must never invent or drop characters around the
// first separator.
func FuzzParseKeyValue(f *testing.F) {
	f.Add(`name=secret`)
	f.Add(`tags=["a","b"]`)
	f.Add(`obj={"k":"v"}`)
	f.Add(`weird==value`)
	f.Add(`broken=[unclosed`)
	f.Add(`noseparator`)
	f.Add(``)
	f.Add(`=emptykey`)

	f.Fuzz(func(t *testing.T, in string) {
		key, val, err := parseKeyValue(in)
		if err != nil {
			return
		}
		if !strings.HasPrefix(in, key+"=") {
			t.Fatalf("parseKeyValue(%q) returned key %q but input does not start with %q=", in, key, key)
		}
		// A literal-string fallback must reproduce the original value half byte
		// for byte; a JSON-decoded value is anything else.
		if s, ok := val.(string); ok {
			if in != key+"="+s {
				t.Fatalf("string value round-trip mismatch: in=%q key=%q val=%q", in, key, s)
			}
		}
	})
}
