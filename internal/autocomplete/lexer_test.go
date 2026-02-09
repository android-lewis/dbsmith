package autocomplete

import (
	"strings"
	"testing"
)

func TestTokenizeDialectFallback(t *testing.T) {
	tokens, err := tokenize("SELECT 1", "unknown")
	if err != nil {
		t.Fatalf("tokenize failed: %v", err)
	}
	if len(tokens) == 0 {
		t.Fatalf("expected tokens")
	}
}

func TestTokenizePositions(t *testing.T) {
	sql := "SELECT\nFROM users"
	tokens, err := tokenize(sql, "postgresql")
	if err != nil {
		t.Fatalf("tokenize failed: %v", err)
	}

	var fromToken *Token
	for i := range tokens {
		trimmed := strings.TrimSpace(tokens[i].Value)
		if strings.EqualFold(trimmed, "FROM") {
			fromToken = &tokens[i]
			break
		}
	}
	if fromToken == nil {
		t.Fatalf("expected FROM token, got %#v", tokens)
	}
	if fromToken.Start.Line != 1 || fromToken.Start.Column != 0 {
		t.Fatalf("unexpected FROM position: %#v", fromToken.Start)
	}
}
