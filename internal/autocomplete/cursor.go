package autocomplete

import (
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
)

var wordMatcher = regexp.MustCompile("([`\"]?[\\w$]+)$")
var qualifierMatcher = regexp.MustCompile(`([A-Za-z_][\w$]*)\.[\w$]*$`)

func currentWord(tokens []Token, sql string, pos Position) (string, Range, string, string, bool) {
	replace := Range{Start: pos, End: pos}
	qualifier := qualifierBeforeCursor(sql, pos)

	token := tokenAtPosition(tokens, pos)
	if token != nil && token.Type.InCategory(chroma.Comment) {
		return "", replace, qualifier, "", true
	}
	if token != nil && token.Type.InCategory(chroma.LiteralString) {
		trimmed := strings.TrimSpace(token.Value)
		if strings.HasPrefix(trimmed, "'") {
			return "", replace, qualifier, "", true
		}
	}

	if token == nil {
		return "", replace, qualifier, "", false
	}

	prefix := prefixToPosition(*token, pos)
	word := lastWord(prefix)
	if word == "" {
		return "", replace, qualifier, "", false
	}

	filter := strings.TrimLeft(word, "`\"")
	quote := ""
	if len(word) > 0 && (word[0] == '`' || word[0] == '"') {
		quote = string(word[0])
	}

	wordStart := trimWordStart(prefix, word)
	startPos := advancePosition(token.Start, wordStart)
	replace = Range{Start: startPos, End: pos}

	return filter, replace, qualifier, quote, false
}

func lastWord(prefix string) string {
	matches := wordMatcher.FindStringSubmatch(prefix)
	if len(matches) > 1 {
		return matches[1]
	}
	if strings.HasSuffix(prefix, "`") || strings.HasSuffix(prefix, "\"") {
		return prefix[len(prefix)-1:]
	}
	return ""
}

func trimWordStart(prefix, word string) string {
	prefixRunes := []rune(prefix)
	wordRunes := []rune(word)
	if len(wordRunes) > len(prefixRunes) {
		return ""
	}
	return string(prefixRunes[:len(prefixRunes)-len(wordRunes)])
}

func tokenAtPosition(tokens []Token, pos Position) *Token {
	for i := range tokens {
		token := &tokens[i]
		if posBefore(pos, token.Start) {
			continue
		}
		if posAfter(pos, token.End) {
			continue
		}
		return token
	}
	return nil
}

func prefixToPosition(token Token, pos Position) string {
	if posBefore(pos, token.Start) {
		return ""
	}
	var builder strings.Builder
	line := token.Start.Line
	col := token.Start.Column

	for _, r := range token.Value {
		if line > pos.Line || (line == pos.Line && col >= pos.Column) {
			break
		}
		builder.WriteRune(r)
		if r == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}

	return builder.String()
}

func qualifierBeforeCursor(sql string, pos Position) string {
	before := textBeforeCursor(sql, pos)
	matches := qualifierMatcher.FindStringSubmatch(before)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func textBeforeCursor(sql string, pos Position) string {
	lines := strings.Split(sql, "\n")
	if pos.Line >= len(lines) {
		pos.Line = len(lines) - 1
		if pos.Line < 0 {
			return ""
		}
	}

	var builder strings.Builder
	for i := 0; i < pos.Line; i++ {
		builder.WriteString(lines[i])
		builder.WriteRune('\n')
	}

	lineRunes := []rune(lines[pos.Line])
	if pos.Column > len(lineRunes) {
		pos.Column = len(lineRunes)
	}
	builder.WriteString(string(lineRunes[:pos.Column]))
	return builder.String()
}

func posBefore(a, b Position) bool {
	if a.Line < b.Line {
		return true
	}
	return a.Line == b.Line && a.Column < b.Column
}

func posAfter(a, b Position) bool {
	if a.Line > b.Line {
		return true
	}
	return a.Line == b.Line && a.Column > b.Column
}
