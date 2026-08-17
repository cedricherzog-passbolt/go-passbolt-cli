package keepass

import (
	"net/url"
	"strings"
	"testing"

	"github.com/tobischo/gokeepasslib/v3"
)

// encodeQuery is a copy of url.Values.Encode that uses %20 instead of '+' for
// spaces, so authenticator apps (Google Authenticator) parse otpauth URIs
// correctly. These tests lock in that distinction.

func TestEncodeQuery_Empty(t *testing.T) {
	if got := encodeQuery(nil); got != "" {
		t.Errorf("nil input = %q, want empty", got)
	}
	if got := encodeQuery(url.Values{}); got != "" {
		t.Errorf("empty input = %q, want empty", got)
	}
}

func TestEncodeQuery_SpacesUsePercent20(t *testing.T) {
	v := url.Values{"issuer": {"My Service"}}
	got := encodeQuery(v)
	if !strings.Contains(got, "%20") {
		t.Errorf("encodeQuery should encode space as %%20, got %q", got)
	}
	if strings.Contains(got, "+") {
		t.Errorf("encodeQuery must not use + for spaces, got %q", got)
	}
}

func TestEncodeQuery_MatchesStdlibAfterPlusSubstitution(t *testing.T) {
	// For values with no plus characters, encodeQuery's output should be
	// identical to url.Values.Encode after replacing '+' with '%20'.
	v := url.Values{
		"issuer": {"Acme"},
		"period": {"30"},
		"digits": {"6"},
	}
	got := encodeQuery(v)
	std := strings.ReplaceAll(v.Encode(), "+", "%20")
	if got != std {
		t.Errorf("encodeQuery(%v) = %q, want %q", v, got, std)
	}
}

func TestEncodeQuery_KeysAreSorted(t *testing.T) {
	v := url.Values{
		"zeta":  {"z"},
		"alpha": {"a"},
		"mu":    {"m"},
	}
	got := encodeQuery(v)
	if !strings.HasPrefix(got, "alpha=a&") || !strings.HasSuffix(got, "&zeta=z") {
		t.Errorf("expected keys in lexicographic order, got %q", got)
	}
}

func TestEncodeQuery_MultiValueRetained(t *testing.T) {
	v := url.Values{"tag": {"foo", "bar"}}
	got := encodeQuery(v)
	if got != "tag=foo&tag=bar" {
		t.Errorf("multi-value encoding = %q, want tag=foo&tag=bar", got)
	}
}

func TestEncodeQuery_DoesNotEscapeAmpersandOrEquals(t *testing.T) {
	// encodeQuery uses url.PathEscape (not QueryEscape) to keep %20 for
	// spaces (required by Google Authenticator). PathEscape does NOT escape
	// '&' or '=', so values containing them produce an ambiguous query
	// string. Acceptable for otpauth labels in practice (they don't contain
	// either character) but a sharp edge — this test locks in current
	// behavior so a future change to QueryEscape is a deliberate decision.
	v := url.Values{"label": {"foo&bar=baz"}}
	got := encodeQuery(v)
	if got != "label=foo&bar=baz" {
		t.Errorf("encodeQuery(%v) = %q, want label=foo&bar=baz (current behavior)", v, got)
	}
}

// addCustomFields delegates the metadata/secret merge to
// helper.ParseCustomFields; these tests cover the KDBX-specific part:
// ordering, the Protected flag, and dropping fields KeePass cannot key.
func TestAddCustomFields(t *testing.T) {
	const idA = "11111111-1111-1111-1111-111111111111"
	const idB = "22222222-2222-2222-2222-222222222222"

	metadata := map[string]any{
		"custom_fields": []any{
			map[string]any{"id": idA, "metadata_key": "api-key"},
			// Cleartext value: the secret half must still carry a (empty)
			// secret_value, so this used to export as blank.
			map[string]any{"id": idB, "metadata_key": "env", "metadata_value": "production"},
		},
	}
	secretFields := map[string]any{
		"custom_fields": []any{
			map[string]any{"id": idA, "secret_value": "sk-secret-123"},
			map[string]any{"id": idB, "secret_value": ""},
		},
	}

	entry := gokeepasslib.NewEntry()
	addCustomFields(&entry, metadata, secretFields)

	want := []struct{ key, value string }{
		{"api-key", "sk-secret-123"},
		{"env", "production"},
	}
	if len(entry.Values) != len(want) {
		t.Fatalf("got %d values, want %d: %+v", len(entry.Values), len(want), entry.Values)
	}
	for i, w := range want {
		got := entry.Values[i]
		if got.Key != w.key {
			t.Errorf("value %d: key = %q, want %q (order must follow the metadata array)", i, got.Key, w.key)
		}
		if got.Value.Content != w.value {
			t.Errorf("value %d (%s): content = %q, want %q", i, w.key, got.Value.Content, w.value)
		}
		if !got.Value.Protected.Bool {
			t.Errorf("value %d (%s): not protected, want every custom field protected", i, w.key)
		}
	}
}

func TestAddCustomFields_SkipsUnnamedAndEmpty(t *testing.T) {
	const idA = "11111111-1111-1111-1111-111111111111"

	// A field with no name on either side cannot be keyed in KDBX.
	entry := gokeepasslib.NewEntry()
	addCustomFields(&entry,
		map[string]any{"custom_fields": []any{map[string]any{"id": idA}}},
		map[string]any{"custom_fields": []any{map[string]any{"id": idA, "secret_value": "v"}}},
	)
	if len(entry.Values) != 0 {
		t.Errorf("unnamed field: got %+v, want no values", entry.Values)
	}

	// A resource without custom fields must add nothing.
	entry = gokeepasslib.NewEntry()
	addCustomFields(&entry, map[string]any{"name": "x"}, map[string]any{"password": "p"})
	if len(entry.Values) != 0 {
		t.Errorf("no custom fields: got %+v, want no values", entry.Values)
	}
}
