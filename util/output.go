package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pterm/pterm"
)

// FprintJSON marshals v with two-space indent and writes it to w followed by a
// newline.
func FprintJSON(w io.Writer, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

// PrintJSON marshals v with two-space indent and prints it to stdout followed by
// a newline. Use for single-entity get output and for full (unfiltered) list
// output.
func PrintJSON(v any) error {
	return FprintJSON(os.Stdout, v)
}

// PrintJSONColumnFiltered marshals each item, re-decodes it into a map, keeps only
// the requested column keys, and prints the indented array to stdout. Use when the
// caller requested an explicit subset of columns; otherwise use PrintJSON(items).
func PrintJSONColumnFiltered[T any](items []T, columns []string) error {
	filtered, err := columnFilteredMaps(items, columns)
	if err != nil {
		return err
	}
	return PrintJSON(filtered)
}

// columnFilteredMaps round-trips each item through JSON and keeps only the entries
// whose key is in columns. The result preserves item order; per-item maps contain
// only the requested keys that exist on the item.
func columnFilteredMaps[T any](items []T, columns []string) ([]map[string]any, error) {
	filtered := make([]map[string]any, len(items))
	for i := range items {
		filtered[i] = make(map[string]any)
		data, err := json.Marshal(items[i])
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("unmarshaling item: %w", err)
		}
		for _, col := range columns {
			if val, ok := m[col]; ok {
				filtered[i][col] = val
			}
		}
	}
	return filtered, nil
}

// PrintTable renders a header row of column names followed by one row per item.
// valueFn returns the rendered cell string for (item, column) and reports whether
// the column is known; an unknown column yields an "unknown column: %q" error.
// It is generic over the item type so it serves both the api.* entity types and
// the resource package's decryptedResource.
func PrintTable[T any](columns []string, items []T, valueFn func(item T, column string) (string, bool)) error {
	data := pterm.TableData{columns}
	for _, it := range items {
		row := make([]string, len(columns))
		for i, col := range columns {
			v, ok := valueFn(it, col)
			if !ok {
				return fmt.Errorf("unknown column: %q", col)
			}
			row[i] = v
		}
		data = append(data, row)
	}
	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	return nil
}
