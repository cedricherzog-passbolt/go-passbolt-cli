package util

import "testing"

// FuzzInitCELProgram throws arbitrary --filter expressions at the cel-go parser
// and compiler. This is the CLI's largest untrusted-text surface; the property is
// simply that compilation never panics and always resolves to exactly one of the
// two valid outcomes — a usable program, or an error — never a nil program with a
// nil error.
func FuzzInitCELProgram(f *testing.F) {
	f.Add(`Name == "foo"`)
	f.Add(`size(Tags) > 0`)
	f.Add(`1 +`)
	f.Add(``)
	f.Add(`"unterminated`)
	f.Add(`undefined_var && true`)

	f.Fuzz(func(t *testing.T, expr string) {
		program, err := InitCELProgram(expr)
		if err == nil && program == nil {
			t.Fatalf("InitCELProgram(%q) returned nil program and nil error", expr)
		}
		if err != nil && program != nil {
			t.Fatalf("InitCELProgram(%q) returned both a program and an error %v", expr, err)
		}
	})
}
