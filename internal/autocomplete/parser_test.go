package autocomplete

import "testing"

func TestAnalyzeKindsSelect(t *testing.T) {
	sql := "SELECT na FROM users"
	pos := Position{Line: 0, Column: len("SELECT na")}

	analysis, err := Analyze(Request{SQL: sql, Position: pos, Dialect: "postgresql"})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if !analysis.HasKind(KindColumn) {
		t.Fatalf("expected column completions")
	}
	if analysis.HasKind(KindTable) {
		t.Fatalf("did not expect table completions in SELECT list")
	}
}

func TestAnalyzeKindsFrom(t *testing.T) {
	sql := "SELECT * FROM us"
	pos := Position{Line: 0, Column: len(sql)}

	analysis, err := Analyze(Request{SQL: sql, Position: pos, Dialect: "postgresql"})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if !analysis.HasKind(KindTable) {
		t.Fatalf("expected table completions after FROM")
	}
}

func TestAnalyzeTableRefs(t *testing.T) {
	sql := "SELECT * FROM public.users u"
	pos := Position{Line: 0, Column: len(sql)}

	analysis, err := Analyze(Request{SQL: sql, Position: pos, Dialect: "postgresql"})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if len(analysis.Tables) != 1 {
		t.Logf("tables: %#v", analysis.Tables)
		t.Fatalf("expected 1 table, got %d", len(analysis.Tables))
	}
	table := analysis.Tables[0]
	if table.Name != "users" || table.Schema != "public" || table.Alias != "u" {
		t.Fatalf("unexpected table ref: %#v", table)
	}
}

func TestAnalyzeInsertColumns(t *testing.T) {
	sql := "INSERT INTO users (na, id) VALUES (1, 2)"
	pos := Position{Line: 0, Column: len("INSERT INTO users (na")}

	analysis, err := Analyze(Request{SQL: sql, Position: pos, Dialect: "postgresql"})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if !analysis.HasKind(KindColumn) {
		t.Fatalf("expected column completions inside INSERT column list")
	}
}
