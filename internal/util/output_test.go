package util

import (
	"bytes"
	"testing"
)

type sampleItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestFprintJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := FprintJSON(&buf, sampleItem{ID: "1", Name: "ada", Age: 30}); err != nil {
		t.Fatalf("FprintJSON: %v", err)
	}
	want := "{\n  \"id\": \"1\",\n  \"name\": \"ada\",\n  \"age\": 30\n}\n"
	if got := buf.String(); got != want {
		t.Errorf("FprintJSON output mismatch:\n got=%q\nwant=%q", got, want)
	}
}

func TestColumnFilteredMaps(t *testing.T) {
	items := []sampleItem{
		{ID: "1", Name: "ada", Age: 30},
		{ID: "2", Name: "grace", Age: 40},
	}

	t.Run("keeps only requested columns", func(t *testing.T) {
		got, err := columnFilteredMaps(items, []string{"id", "name"})
		if err != nil {
			t.Fatalf("columnFilteredMaps: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(got))
		}
		for i, m := range got {
			if _, ok := m["age"]; ok {
				t.Errorf("entry %d should not contain dropped column %q", i, "age")
			}
			if len(m) != 2 {
				t.Errorf("entry %d expected 2 keys, got %d: %v", i, len(m), m)
			}
		}
		if got[0]["id"] != "1" || got[0]["name"] != "ada" {
			t.Errorf("entry 0 unexpected values: %v", got[0])
		}
	})

	t.Run("unknown column is ignored", func(t *testing.T) {
		got, err := columnFilteredMaps(items, []string{"id", "nonexistent"})
		if err != nil {
			t.Fatalf("columnFilteredMaps: %v", err)
		}
		if len(got[0]) != 1 {
			t.Errorf("expected only the existing column to be kept, got %v", got[0])
		}
	})

	t.Run("empty items yields empty slice", func(t *testing.T) {
		got, err := columnFilteredMaps([]sampleItem{}, []string{"id"})
		if err != nil {
			t.Fatalf("columnFilteredMaps: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected empty result, got %v", got)
		}
	})
}

func TestPrintTable_UnknownColumn(t *testing.T) {
	items := []sampleItem{{ID: "1", Name: "ada"}}

	valueFn := func(it sampleItem, col string) (string, bool) {
		switch col {
		case "id":
			return it.ID, true
		case "name":
			return it.Name, true
		default:
			return "", false
		}
	}

	// Known columns render without error.
	if err := PrintTable([]string{"id", "name"}, items, valueFn); err != nil {
		t.Fatalf("PrintTable with known columns: %v", err)
	}

	// An unknown column returns the standard error.
	err := PrintTable([]string{"id", "bogus"}, items, valueFn)
	if err == nil {
		t.Fatal("expected error for unknown column, got nil")
	}
	if want := `unknown column: "bogus"`; err.Error() != want {
		t.Errorf("unexpected error: got %q, want %q", err.Error(), want)
	}
}

func TestPrintTable_ValueFnError(t *testing.T) {
	// Guard the generic constraint compiles with a non-struct item type too.
	items := []string{"a", "b"}
	calls := 0
	valueFn := func(it string, col string) (string, bool) {
		calls++
		return it, true
	}
	if err := PrintTable([]string{"col"}, items, valueFn); err != nil {
		t.Fatalf("PrintTable: %v", err)
	}
	if calls != len(items) {
		t.Errorf("expected valueFn called %d times, got %d", len(items), calls)
	}
}
