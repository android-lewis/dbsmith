package autocomplete

import "testing"

func TestCompleteKeywords(t *testing.T) {
	result, err := Complete(Request{
		SQL: "sel",
		Position: Position{
			Line:   0,
			Column: 3,
		},
		Dialect: "postgresql",
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}

	if !hasItem(result.Items, "SELECT", KindKeyword) {
		t.Fatalf("expected SELECT keyword, got %#v", result.Items)
	}
}

func TestCompleteTables(t *testing.T) {
	result, err := Complete(Request{
		SQL: "SELECT * FROM us",
		Position: Position{
			Line:   0,
			Column: len("SELECT * FROM us"),
		},
		Dialect: "postgresql",
		Context: Context{
			Tables: []Table{
				{Name: "users"},
				{Name: "orders"},
			},
		},
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}

	if !hasItem(result.Items, "users", KindTable) {
		t.Fatalf("expected users table, got %#v", result.Items)
	}
}

func TestCompleteColumns(t *testing.T) {
	result, err := Complete(Request{
		SQL: "SELECT na FROM users",
		Position: Position{
			Line:   0,
			Column: len("SELECT na"),
		},
		Dialect: "postgresql",
		Context: Context{
			Tables: []Table{
				{
					Name: "users",
					Columns: []Column{
						{Name: "id"},
						{Name: "name"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}

	if !hasItem(result.Items, "name", KindColumn) {
		t.Fatalf("expected name column, got %#v", result.Items)
	}
}

func hasItem(items []Item, label string, kind ItemKind) bool {
	for _, item := range items {
		if item.Label == label && item.Kind == kind {
			return true
		}
	}
	return false
}
