package autocomplete

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

type Token struct {
	Type  chroma.TokenType
	Value string
	Start Position
	End   Position
}

func tokenize(sql, dialect string) ([]Token, error) {
	lexerName := mapDialectToLexer(dialect)
	lexer := lexers.Get(lexerName)
	if lexer == nil {
		lexer = lexers.Get("sql")
	}
	if lexer == nil {
		return nil, ErrLexerNotFound
	}

	iter, err := lexer.Tokenise(nil, sql)
	if err != nil {
		return nil, err
	}

	var tokens []Token
	line := 0
	col := 0

	for _, token := range iter.Tokens() {
		if token.Value == "" {
			continue
		}

		start := Position{Line: line, Column: col}
		end := advancePosition(start, token.Value)

		tokens = append(tokens, Token{
			Type:  token.Type,
			Value: token.Value,
			Start: start,
			End:   end,
		})

		line = end.Line
		col = end.Column
	}

	return tokens, nil
}

func mapDialectToLexer(dialect string) string {
	switch strings.ToLower(dialect) {
	case "postgres", "postgresql":
		return "postgresql"
	case "mysql":
		return "mysql"
	case "sqlite", "sqlite3":
		return "sqlite3"
	default:
		return "sql"
	}
}

func advancePosition(start Position, text string) Position {
	line := start.Line
	col := start.Column
	for _, r := range text {
		if r == '\n' {
			line++
			col = 0
			continue
		}
		col++
	}
	return Position{Line: line, Column: col}
}
