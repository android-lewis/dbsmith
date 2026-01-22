package syntax

import (
	"errors"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/gdamore/tcell/v2"
)

var (
	ErrLexerNotFound = errors.New("lexer not found")
)

type StyledToken struct {
	Text  string
	Line  int
	Col   int
	Style tcell.Style
}

type Highlighter struct {
	theme      *Theme
	lexerCache map[string]chroma.Lexer
	cacheMutex sync.RWMutex
}

func NewHighlighter(theme *Theme) *Highlighter {
	return &Highlighter{
		theme:      theme,
		lexerCache: make(map[string]chroma.Lexer),
	}
}

func (h *Highlighter) Highlight(sql string, dialect string) ([]StyledToken, error) {
	lexer, err := h.getLexer(dialect)
	if err != nil {
		return h.plainTextFallback(sql), nil
	}

	iterator, err := lexer.Tokenise(nil, sql)
	if err != nil {
		return h.plainTextFallback(sql), nil
	}

	var tokens []StyledToken
	line := 0
	col := 0

	for _, token := range iterator.Tokens() {
		style := h.theme.ChromaToTcell(token)

		lines := strings.Split(token.Value, "\n")
		for i, lineText := range lines {
			if i > 0 {
				line++
				col = 0
			}

			if lineText != "" {
				tokens = append(tokens, StyledToken{
					Text:  lineText,
					Line:  line,
					Col:   col,
					Style: style,
				})
				col += len(lineText)
			}
		}
	}

	return tokens, nil
}

func (h *Highlighter) SetTheme(theme *Theme) {
	h.theme = theme
}

func (h *Highlighter) getLexer(dialect string) (chroma.Lexer, error) {
	h.cacheMutex.RLock()
	lexer, exists := h.lexerCache[dialect]
	h.cacheMutex.RUnlock()

	if exists {
		return lexer, nil
	}

	lexerName := h.mapDialectToLexer(dialect)

	lexer = lexers.Get(lexerName)
	if lexer == nil {
		lexer = lexers.Get("sql")
	}

	if lexer == nil {
		return nil, ErrLexerNotFound
	}

	h.cacheMutex.Lock()
	h.lexerCache[dialect] = lexer
	h.cacheMutex.Unlock()

	return lexer, nil
}

func (h *Highlighter) mapDialectToLexer(dialect string) string {
	switch strings.ToLower(dialect) {
	case "postgres", "postgresql":
		return "postgresql"
	case "mysql":
		return "mysql"
	case "tsql", "mssql", "sqlserver":
		return "tsql"
	case "sqlite", "sqlite3":
		return "sqlite3"
	default:
		return "sql"
	}
}

func (h *Highlighter) plainTextFallback(sql string) []StyledToken {
	defaultStyle := h.theme.GetStyle(chroma.Text)
	lines := strings.Split(sql, "\n")

	var tokens []StyledToken
	for lineNum, lineText := range lines {
		if lineText != "" {
			tokens = append(tokens, StyledToken{
				Text:  lineText,
				Line:  lineNum,
				Col:   0,
				Style: defaultStyle,
			})
		}
	}

	return tokens
}
