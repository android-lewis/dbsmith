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
	Tables         []Table
	ColumnsByTable map[string][]Column
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

type Analysis struct {
	Word      string
	Replace   Range
	Qualifier string
	Quote     string
	Kinds     []ItemKind
	Tables    []TableRef
	Aliases   map[string]string
	Suppress  bool
}

func (a Analysis) HasKind(kind ItemKind) bool {
	for _, k := range a.Kinds {
		if k == kind {
			return true
		}
	}
	return false
}

func (a Analysis) TargetTables() []string {
	stmt := statement{
		Tables:      a.Tables,
		Aliases:     a.Aliases,
		TableLookup: tableLookupFromRefs(a.Tables),
	}
	if a.Qualifier != "" {
		if target := resolveQualifiedTable(a.Qualifier, stmt); target != "" {
			return []string{target}
		}
	}
	if len(a.Tables) == 0 {
		return nil
	}
	return dedupeTableNames(a.Tables)
}

func Complete(req Request) (Result, error) {
	analysis, err := Analyze(req)
	if err != nil {
		return Result{}, err
	}
	return CompleteWithAnalysis(analysis, req.Context, req.Dialect), nil
}

func Analyze(req Request) (Analysis, error) {
	tokens, err := tokenize(req.SQL, req.Dialect)
	if err != nil {
		return Analysis{}, err
	}

	word, replaceRange, qualifier, quote, suppress := currentWord(tokens, req.SQL, req.Position)
	if suppress {
		return Analysis{
			Replace:   replaceRange,
			Word:      word,
			Qualifier: qualifier,
			Quote:     quote,
			Suppress:  true,
		}, nil
	}

	lexemes := buildLexemes(tokens)
	stmt := parseStatement(lexemes)
	kinds := detectCompletionKinds(lexemes, req.Position)

	return Analysis{
		Word:      word,
		Replace:   replaceRange,
		Qualifier: qualifier,
		Quote:     quote,
		Kinds:     kinds.kinds,
		Tables:    stmt.Tables,
		Aliases:   stmt.Aliases,
	}, nil
}

func CompleteWithAnalysis(analysis Analysis, context Context, dialect string) Result {
	if analysis.Suppress {
		return Result{Replace: analysis.Replace}
	}

	ctx := completionContext{kinds: analysis.Kinds}
	stmt := statement{
		Tables:      analysis.Tables,
		Aliases:     analysis.Aliases,
		TableLookup: tableLookupFromRefs(analysis.Tables),
	}

	items := buildCandidates(ctx, stmt, context, dialect, analysis.Qualifier, analysis.Quote)
	items = filterCandidates(items, analysis.Word)
	sortCandidates(items)

	return Result{
		Items:   items,
		Replace: analysis.Replace,
	}
}

func sortCandidates(items []Item) {
	if len(items) <= 1 {
		return
	}
	ranks := make([]int, len(items))
	labels := make([]string, len(items))
	for i, item := range items {
		ranks[i] = kindRank(item.Kind)
		labels[i] = strings.ToUpper(item.Label)
	}
	sort.SliceStable(items, func(i, j int) bool { //TODO: use slices.SortStableFunc
		if ranks[i] != ranks[j] {
			return ranks[i] < ranks[j]
		}
		return labels[i] < labels[j]
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

func dedupeTableNames(tables []TableRef) []string {
	seen := map[string]bool{}
	results := make([]string, 0, len(tables))
	for _, table := range tables {
		key := strings.ToUpper(table.Name)
		if seen[key] {
			continue
		}
		seen[key] = true
		results = append(results, table.Name)
	}
	return results
}
