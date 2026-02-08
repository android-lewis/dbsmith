package autocomplete

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
)

type lexemeKind int

const (
	lexemeKeyword lexemeKind = iota
	lexemeIdentifier
	lexemePunctuation
)

type lexeme struct {
	Kind  lexemeKind
	Value string
	Start Position
	End   Position
}

type tableRef struct {
	Name   string
	Schema string
	Alias  string
}

type statement struct {
	Tables  []tableRef
	Aliases map[string]string
}

type completionContext struct {
	kinds []ItemKind
}

type clauseKind int

const (
	clauseUnknown clauseKind = iota
	clauseSelect
	clauseFrom
	clauseWhere
	clauseJoin
	clauseUpdate
	clauseInto
	clauseDelete
	clauseSet
	clauseGroup
	clauseOrder
	clauseHaving
)

func buildLexemes(tokens []Token) []lexeme {
	lexemes := make([]lexeme, 0, len(tokens))
	for _, token := range tokens {
		if token.Type.InCategory(chroma.Comment) {
			continue
		}

		trimmed := strings.TrimSpace(token.Value)
		if trimmed == "" {
			continue
		}

		upper := strings.ToUpper(trimmed)
		switch {
		case isKeyword(upper):
			lexemes = append(lexemes, lexeme{
				Kind:  lexemeKeyword,
				Value: upper,
				Start: token.Start,
				End:   token.End,
			})
		case token.Type.InCategory(chroma.Keyword):
			lexemes = append(lexemes, lexeme{
				Kind:  lexemeKeyword,
				Value: upper,
				Start: token.Start,
				End:   token.End,
			})
		case token.Type.InCategory(chroma.Name) || token.Type.InCategory(chroma.LiteralString):
			lexemes = append(lexemes, lexeme{
				Kind:  lexemeIdentifier,
				Value: normalizeIdentifier(trimmed),
				Start: token.Start,
				End:   token.End,
			})
		case token.Type.InCategory(chroma.Punctuation) || token.Type.InCategory(chroma.Operator):
			lexemes = append(lexemes, lexeme{
				Kind:  lexemePunctuation,
				Value: trimmed,
				Start: token.Start,
				End:   token.End,
			})
		}
	}
	return lexemes
}

func parseStatement(lexemes []lexeme) statement {
	tables, aliases := extractTables(lexemes)
	return statement{
		Tables:  tables,
		Aliases: aliases,
	}
}

func extractTables(lexemes []lexeme) ([]tableRef, map[string]string) {
	var tables []tableRef
	aliases := map[string]string{}

	for i := 0; i < len(lexemes); i++ {
		lex := lexemes[i]
		if lex.Kind != lexemeKeyword || !isTableStartKeyword(lex.Value) {
			continue
		}

		refs, next := parseTablesAfter(lexemes, i+1)
		for _, ref := range refs {
			tables = append(tables, ref)
			if ref.Alias != "" {
				aliases[strings.ToUpper(ref.Alias)] = ref.Name
			}
		}
		i = next
	}

	return tables, aliases
}

func parseTablesAfter(lexemes []lexeme, start int) ([]tableRef, int) {
	var refs []tableRef
	i := start
	for i < len(lexemes) {
		lex := lexemes[i]
		if lex.Kind == lexemeKeyword && isTableStopKeyword(lex.Value) {
			break
		}
		if lex.Kind == lexemePunctuation && strings.Contains(lex.Value, ",") {
			i++
			continue
		}
		if lex.Kind != lexemeIdentifier {
			i++
			continue
		}

		ref, next := parseSingleTable(lexemes, i)
		refs = append(refs, ref)
		i = next + 1
	}
	return refs, i - 1
}

func parseSingleTable(lexemes []lexeme, start int) (tableRef, int) {
	name := lexemes[start].Value
	schema := ""
	i := start + 1

	if i+1 < len(lexemes) &&
		lexemes[i].Kind == lexemePunctuation &&
		lexemes[i].Value == "." &&
		lexemes[i+1].Kind == lexemeIdentifier {
		schema = name
		name = lexemes[i+1].Value
		i += 2
	}

	alias := ""
	if i < len(lexemes) && lexemes[i].Kind == lexemeKeyword && lexemes[i].Value == "AS" {
		if i+1 < len(lexemes) && lexemes[i+1].Kind == lexemeIdentifier {
			alias = lexemes[i+1].Value
			i += 2
		}
	} else if i < len(lexemes) && lexemes[i].Kind == lexemeIdentifier {
		alias = lexemes[i].Value
		i++
	}

	return tableRef{
		Name:   name,
		Schema: schema,
		Alias:  alias,
	}, i - 1
}

func detectCompletionKinds(lexemes []lexeme, pos Position) completionContext {
	insertRange := findInsertColumnsRange(lexemes)
	if insertRange != nil && posInRange(pos, insertRange.Start, insertRange.End) {
		return completionContext{kinds: []ItemKind{KindColumn, KindKeyword}}
	}

	depth := 0
	clause := clauseUnknown

	for _, lex := range lexemes {
		if !posAfterStart(pos, lex.Start) {
			break
		}

		if lex.Kind == lexemePunctuation {
			depth += strings.Count(lex.Value, "(")
			depth -= strings.Count(lex.Value, ")")
			if depth < 0 {
				depth = 0
			}
			continue
		}

		if depth > 0 || lex.Kind != lexemeKeyword {
			continue
		}

		switch lex.Value {
		case "SELECT":
			clause = clauseSelect
		case "FROM":
			clause = clauseFrom
		case "WHERE":
			clause = clauseWhere
		case "JOIN":
			clause = clauseJoin
		case "UPDATE":
			clause = clauseUpdate
		case "INTO":
			clause = clauseInto
		case "DELETE":
			clause = clauseDelete
		case "SET":
			clause = clauseSet
		case "GROUP":
			clause = clauseGroup
		case "ORDER":
			clause = clauseOrder
		case "HAVING":
			clause = clauseHaving
		}
	}

	kinds := []ItemKind{KindKeyword}
	switch clause {
	case clauseSelect, clauseWhere, clauseSet, clauseGroup, clauseOrder, clauseHaving:
		kinds = append(kinds, KindColumn)
	case clauseFrom, clauseJoin, clauseUpdate, clauseInto, clauseDelete:
		kinds = append(kinds, KindTable)
	}

	return completionContext{kinds: kinds}
}

func findInsertColumnsRange(lexemes []lexeme) *Range {
	const (
		insertNone = iota
		insertSeen
		insertInto
		insertAfterTable
		insertInColumns
	)

	state := insertNone
	depth := 0
	var start Position

	for _, lex := range lexemes {
		switch lex.Kind {
		case lexemeKeyword:
			switch lex.Value {
			case "INSERT":
				state = insertSeen
			case "INTO":
				if state == insertSeen {
					state = insertInto
				}
			case "VALUES":
				if state == insertInColumns {
					return &Range{Start: start, End: lex.Start}
				}
				state = insertNone
			}
		case lexemeIdentifier:
			if state == insertInto {
				state = insertAfterTable
			}
		case lexemePunctuation:
			if lex.Value == "(" {
				if state == insertAfterTable && depth == 0 {
					start = lex.End
					state = insertInColumns
				}
				depth++
			} else if lex.Value == ")" {
				depth--
				if depth < 0 {
					depth = 0
				}
				if state == insertInColumns && depth == 0 {
					return &Range{Start: start, End: lex.Start}
				}
			}
		}
	}

	return nil
}

func isTableStartKeyword(keyword string) bool {
	switch keyword {
	case "FROM", "JOIN", "UPDATE", "INTO", "DELETE":
		return true
	default:
		return false
	}
}

func isTableStopKeyword(keyword string) bool {
	switch keyword {
	case "WHERE", "GROUP", "ORDER", "HAVING", "JOIN", "SET", "VALUES", "LIMIT", "OFFSET":
		return true
	default:
		return false
	}
}

func normalizeIdentifier(value string) string {
	return strings.Trim(value, "`\"")
}

func posAfterStart(pos Position, start Position) bool {
	if pos.Line > start.Line {
		return true
	}
	if pos.Line == start.Line && pos.Column >= start.Column {
		return true
	}
	return false
}

func posInRange(pos, start, end Position) bool {
	if pos.Line < start.Line || (pos.Line == start.Line && pos.Column < start.Column) {
		return false
	}
	if pos.Line > end.Line || (pos.Line == end.Line && pos.Column > end.Column) {
		return false
	}
	return true
}
