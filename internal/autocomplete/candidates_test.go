package autocomplete

import (
	"slices"
	"strings"
	"testing"
)

func TestColumnCandidatesWithQualifier(t *testing.T) {
	ctx := Context{
		Tables: []Table{
			{Name: "users", Columns: []Column{{Name: "id"}, {Name: "name"}}},
			{Name: "orders", Columns: []Column{{Name: "id"}, {Name: "total"}}},
		},
		ColumnsByTable: map[string][]Column{
			"USERS":  {{Name: "id"}, {Name: "name"}},
			"ORDERS": {{Name: "id"}, {Name: "total"}},
		},
	}

	stmt := statement{
		Tables: []TableRef{
			{Name: "users", Alias: "u"},
			{Name: "orders", Alias: "o"},
		},
		Aliases:     map[string]string{"U": "users", "O": "orders"},
		TableLookup: tableLookupFromRefs([]TableRef{{Name: "users"}, {Name: "orders"}}),
	}

	items := columnCandidates(stmt, ctx, "u", "")
	if len(items) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(items))
	}
	for _, item := range items {
		if !strings.Contains(item.Detail, "users") {
			t.Fatalf("expected users column, got %#v", item)
		}
	}
}

func TestFilterCandidatesWithSchemaPrefix(t *testing.T) {
	items := []Item{
		{Label: "public.users", Kind: KindTable},
		{Label: "orders", Kind: KindTable},
	}
	filtered := filterCandidates(items, "us")
	if len(filtered) != 1 || filtered[0].Label != "public.users" {
		t.Fatalf("unexpected filter result: %#v", filtered)
	}
}

func TestCandidateOrderingByKindThenName(t *testing.T) {
	ctx := Context{
		Tables: []Table{
			{Name: "users", Columns: []Column{{Name: "name"}, {Name: "id"}}},
		},
		ColumnsByTable: map[string][]Column{
			"USERS": {{Name: "name"}, {Name: "id"}},
		},
	}
	analysis := Analysis{
		Kinds: []ItemKind{KindColumn, KindTable, KindKeyword},
	}
	result := CompleteWithAnalysis(analysis, ctx, "postgresql")
	if len(result.Items) == 0 {
		t.Fatalf("expected candidates")
	}

	kinds := make([]ItemKind, 0, len(result.Items))
	for _, item := range result.Items {
		kinds = append(kinds, item.Kind)
	}

	firstTable := slices.Index(kinds, KindTable)
	firstKeyword := slices.Index(kinds, KindKeyword)
	if firstTable == -1 || firstKeyword == -1 {
		t.Fatalf("missing table or keyword candidates: %#v", kinds)
	}
	for i := 0; i < firstTable; i++ {
		if kinds[i] != KindColumn {
			t.Fatalf("expected columns first, got kind %v at %d", kinds[i], i)
		}
	}
	for i := firstTable; i < firstKeyword; i++ {
		if kinds[i] != KindTable {
			t.Fatalf("expected tables before keywords, got kind %v at %d", kinds[i], i)
		}
	}

	var columnLabels []string
	for _, item := range result.Items {
		if item.Kind == KindColumn {
			columnLabels = append(columnLabels, item.Label)
		}
	}
	if !slices.IsSortedFunc(columnLabels, func(a, b string) int {
		return strings.Compare(strings.ToUpper(a), strings.ToUpper(b))
	}) {
		t.Fatalf("expected columns sorted alphabetically: %#v", columnLabels)
	}
}

func TestCompleteWithAliasQualifierPrefix(t *testing.T) {
	sql := "SELECT c.p FROM public.comments as c"
	pos := Position{Line: 0, Column: len("SELECT c.p")}

	analysis, err := Analyze(Request{SQL: sql, Position: pos, Dialect: "postgresql"})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if !analysis.HasKind(KindColumn) {
		t.Fatalf("expected column completions")
	}

	context := Context{
		Tables: []Table{
			{Name: "comments", Schema: "public", Columns: []Column{{Name: "post_id"}, {Name: "body"}}},
		},
		ColumnsByTable: map[string][]Column{
			"COMMENTS": {{Name: "post_id"}, {Name: "body"}},
		},
	}

	result := CompleteWithAnalysis(analysis, context, "postgresql")
	if len(result.Items) == 0 {
		t.Fatalf("expected completions, got none")
	}
	found := false
	for _, item := range result.Items {
		if item.Kind == KindColumn && item.Label == "post_id" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected post_id completion, got %#v", result.Items)
	}
}
