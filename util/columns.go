package util

import (
	"fmt"
	"strings"
)

// ColumnAlias declares a canonical column name together with the legacy input
// forms that should resolve to it. Used by list commands to accept
// snake_case (canonical), PascalCase, and lowercased-concat forms uniformly
// in --column and --filter flags.
type ColumnAlias struct {
	Name    string
	Aliases []string
}

// ColumnAliasResolver normalizes user-supplied column names to their
// canonical form via case-insensitive lookup against name+aliases.
type ColumnAliasResolver struct {
	canonical []string
	lookup    map[string]string
}

func NewColumnAliasResolver(specs []ColumnAlias) *ColumnAliasResolver {
	r := &ColumnAliasResolver{
		canonical: make([]string, 0, len(specs)),
		lookup:    make(map[string]string),
	}
	for _, s := range specs {
		r.canonical = append(r.canonical, s.Name)
		r.lookup[strings.ToLower(s.Name)] = s.Name
		for _, a := range s.Aliases {
			r.lookup[strings.ToLower(a)] = s.Name
		}
	}
	return r
}

// Normalize maps an input column name to its canonical form. Returns an
// error listing the valid canonical names if the input matches nothing.
func (r *ColumnAliasResolver) Normalize(input string) (string, error) {
	if canon, ok := r.lookup[strings.ToLower(input)]; ok {
		return canon, nil
	}
	return "", fmt.Errorf("unknown column: %q (valid: %s)", input, strings.Join(r.canonical, ", "))
}

// NormalizeAll normalizes a slice of input column names, returning the
// canonical-name slice or the first error encountered.
func (r *ColumnAliasResolver) NormalizeAll(inputs []string) ([]string, error) {
	out := make([]string, len(inputs))
	for i, in := range inputs {
		canon, err := r.Normalize(in)
		if err != nil {
			return nil, err
		}
		out[i] = canon
	}
	return out, nil
}

// Canonical returns the canonical column names in declaration order. Useful
// for building --column flag help text and default value slices.
func (r *ColumnAliasResolver) Canonical() []string {
	out := make([]string, len(r.canonical))
	copy(out, r.canonical)
	return out
}
