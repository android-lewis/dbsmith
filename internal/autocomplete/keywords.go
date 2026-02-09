package autocomplete

import "strings"

var baseKeywords = []string{
	"SELECT",
	"FROM",
	"WHERE",
	"INSERT",
	"UPDATE",
	"DELETE",
	"INTO",
	"VALUES",
	"JOIN",
	"LEFT",
	"RIGHT",
	"INNER",
	"OUTER",
	"FULL",
	"GROUP",
	"ORDER",
	"BY",
	"HAVING",
	"LIMIT",
	"OFFSET",
	"SET",
	"AS",
	"AND",
	"OR",
	"NOT",
	"NULL",
	"IS",
	"IN",
	"EXISTS",
	"DISTINCT",
	"CASE",
	"WHEN",
	"THEN",
	"ELSE",
	"END",
}

var postgresKeywords = []string{
	"RETURNING",
	"ILIKE",
}

var mysqlKeywords = []string{
	"REPLACE",
	"SHOW",
}

var sqliteKeywords = []string{
	"PRAGMA",
	"EXPLAIN",
}

func keywordsForDialect(dialect string) []string {
	keywords := append([]string{}, baseKeywords...)
	switch strings.ToLower(dialect) {
	case "postgres", "postgresql":
		keywords = append(keywords, postgresKeywords...)
	case "mysql":
		keywords = append(keywords, mysqlKeywords...)
	case "sqlite", "sqlite3":
		keywords = append(keywords, sqliteKeywords...)
	}
	return dedupeKeywords(keywords)
}

func dedupeKeywords(keywords []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		upper := strings.ToUpper(keyword)
		if seen[upper] {
			continue
		}
		seen[upper] = true
		result = append(result, upper)
	}
	return result
}

func isKeyword(keyword string) bool {
	_, ok := keywordLookup[keyword]
	return ok
}

var keywordLookup = keywordSet()

func keywordSet() map[string]bool {
	set := make(map[string]bool)
	for _, keyword := range keywordsForDialect("") {
		set[keyword] = true
	}
	for _, keyword := range postgresKeywords {
		set[strings.ToUpper(keyword)] = true
	}
	for _, keyword := range mysqlKeywords {
		set[strings.ToUpper(keyword)] = true
	}
	for _, keyword := range sqliteKeywords {
		set[strings.ToUpper(keyword)] = true
	}
	return set
}
