package user

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/passbolt/go-passbolt/api"
)

// UserJSONOutput's lifecycle fields (active, deleted, disabled) must emit
// even when false so that --filter 'active == false' isn't blocked by
// omitempty-style stripping. Regression guard.
func TestUserJSONOutput_LifecycleFieldsAlwaysEmitted(t *testing.T) {
	out := UserJSONOutput{
		ID:       ptr("u1"),
		Username: ptr("ada@passbolt.com"),
		Active:   false,
		Deleted:  false,
		Disabled: false,
	}
	got, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"active":false`, `"deleted":false`, `"disabled":false`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in output %s", want, got)
		}
	}
}

// printJSONUsers must map api.User onto UserJSONOutput with the lifecycle
// flags correctly derived — Active/Deleted as direct passthrough, Disabled
// from User.Disabled != nil. Catches struct-literal-side bugs that the JSON
// tag test above can't.
func TestPrintJSONUsers_PopulatesLifecycleFields(t *testing.T) {
	disabledTime := api.Time{}
	users := []api.User{
		{
			ID: "active-user", Username: "ada@passbolt.com",
			Profile: &api.Profile{FirstName: "Ada", LastName: "Lovelace"},
			Role:    &api.Role{Name: "user"},
			Created: &api.Time{}, Modified: &api.Time{},
			Active: true, Deleted: false, Disabled: nil,
		},
		{
			ID: "disabled-user", Username: "grace@passbolt.com",
			Profile: &api.Profile{FirstName: "Grace", LastName: "Hopper"},
			Role:    &api.Role{Name: "user"},
			Created: &api.Time{}, Modified: &api.Time{},
			Active: true, Deleted: false, Disabled: &disabledTime,
		},
		{
			ID: "deleted-user", Username: "old@passbolt.com",
			Profile: &api.Profile{FirstName: "Old", LastName: "Account"},
			Role:    &api.Role{Name: "user"},
			Created: &api.Time{}, Modified: &api.Time{},
			Active: false, Deleted: true, Disabled: nil,
		},
	}

	out := captureStdout(t, func() {
		if err := printJSONUsers(users, false, nil); err != nil {
			t.Fatalf("printJSONUsers: %v", err)
		}
	})

	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v\nraw:\n%s", err, out)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 users in output, got %d", len(got))
	}

	cases := []struct {
		id              string
		active, deleted bool
		disabled        bool
	}{
		{"active-user", true, false, false},
		{"disabled-user", true, false, true},
		{"deleted-user", false, true, false},
	}
	for _, c := range cases {
		var record map[string]any
		for _, r := range got {
			if r["id"] == c.id {
				record = r
				break
			}
		}
		if record == nil {
			t.Errorf("no record for id=%q", c.id)
			continue
		}
		if record["active"] != c.active {
			t.Errorf("%s: active = %v, want %v", c.id, record["active"], c.active)
		}
		if record["deleted"] != c.deleted {
			t.Errorf("%s: deleted = %v, want %v", c.id, record["deleted"], c.deleted)
		}
		if record["disabled"] != c.disabled {
			t.Errorf("%s: disabled = %v, want %v", c.id, record["disabled"], c.disabled)
		}
	}
}

func ptr[T any](v T) *T { return &v }

// captureStdout swaps os.Stdout for a pipe, runs fn, restores stdout, and
// returns whatever fn wrote. Used to test functions that print directly
// rather than returning an io.Writer-bound result.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()
	_ = w.Close()
	<-done
	return buf.String()
}
