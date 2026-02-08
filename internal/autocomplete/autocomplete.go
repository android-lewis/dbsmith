package autocomplete

import (
	"errors"
	"sort"
	"strings"
)

var ErrLexerNotFound = errors.New("lexer not found")

type Position struct {
	Line   int
	Column int
}

type Range struct {
	Start Position
	End   Position
}

type Column struct {
	Name string
}

type Table struct {
	Name    string
	Schema  string
	Columns []Column
}

type Context struct {
	Tables []Table
}

type ItemKind int

const (
	KindKeyword ItemKind = iota
	KindTable
	KindColumn
)

type Item struct {
	Label  string
	Kind   ItemKind
	Detail string
}

type Request struct {
	SQL      string
	Position Position
	Dialect  string
	Context  Context
}

type Result struct {
	Items   []Item
	Replace Range
}

func Complete(req Request) (Result, error) {
	tokens, err := tokenize(req.SQL, req.Dialect)
	if err != nil {
		return Result{}, err
	}

	word, replaceRange, qualifier, quote, suppress := currentWord(tokens, req.SQL, req.Position)
	if suppress {
		return Result{Replace: replaceRange}, nil
	}

	lexemes := buildLexemes(tokens)
	stmt := parseStatement(lexemes)
	kinds := detectCompletionKinds(lexemes, req.Position)

	items := buildCandidates(kinds, stmt, req.Context, req.Dialect, qualifier, quote)
	items = filterCandidates(items, word)
	sortCandidates(items)

	return Result{
		Items:   items,
		Replace: replaceRange,
	}, nil
}

func sortCandidates(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		rankI := kindRank(items[i].Kind)
		rankJ := kindRank(items[j].Kind)
		if rankI != rankJ {
			return rankI < rankJ
		}
		return strings.ToUpper(items[i].Label) < strings.ToUpper(items[j].Label)
	})
}

func kindRank(kind ItemKind) int {
	switch kind {
	case KindColumn:
		return 0
	case KindTable:
		return 1
	case KindKeyword:
		return 2
	default:
		return 3
	}
}
