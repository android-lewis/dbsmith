package autocomplete

import (
	"testing"

	"github.com/alecthomas/chroma/v2"
)

func TestAnalyzeCurrentWordAndReplace(t *testing.T) {
	sql := "SELECT na FROM users"
	pos := Position{Line: 0, Column: len("SELECT na")}

	analysis, err := Analyze(Request{SQL: sql, Position: pos, Dialect: "postgresql"})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if analysis.Word != "na" {
		t.Fatalf("expected word 'na', got %q", analysis.Word)
	}
	if analysis.Replace.Start.Column != len("SELECT ") {
		t.Fatalf("unexpected replace start: %#v", analysis.Replace)
	}
	if analysis.Replace.End.Column != pos.Column {
		t.Fatalf("unexpected replace end: %#v", analysis.Replace)
	}
}

func TestAnalyzeQualifierWithoutWord(t *testing.T) {
	sql := "SELECT u."
	pos := Position{Line: 0, Column: len(sql)}

	analysis, err := Analyze(Request{SQL: sql, Position: pos, Dialect: "postgresql"})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if analysis.Word != "" {
		t.Fatalf("expected empty word, got %q", analysis.Word)
	}
	if analysis.Qualifier != "u" {
		t.Fatalf("expected qualifier u, got %q", analysis.Qualifier)
	}
}

func TestAnalyzeSuppressInString(t *testing.T) {
	sql := "SELECT 'name'"
	tokens, err := tokenize(sql, "postgresql")
	if err != nil {
		t.Fatalf("tokenize failed: %v", err)
	}
	var stringToken *Token
	for i := range tokens {
		if tokens[i].Type.InCategory(chroma.LiteralString) {
			stringToken = &tokens[i]
			break
		}
	}
	if stringToken == nil {
		t.Fatalf("expected string token, got %#v", tokens)
	}
	pos := Position{Line: stringToken.Start.Line, Column: stringToken.Start.Column + 1}

	analysis, err := Analyze(Request{SQL: sql, Position: pos, Dialect: "postgresql"})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if !analysis.Suppress {
		t.Fatalf("expected suppression inside string")
	}
}

func TestAnalyzeSuppressInComment(t *testing.T) {
	sql := "-- comment\nSELECT"
	pos := Position{Line: 0, Column: 3}

	analysis, err := Analyze(Request{SQL: sql, Position: pos, Dialect: "postgresql"})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if !analysis.Suppress {
		t.Fatalf("expected suppression inside comment")
	}
}
