package editor

import (
	"regexp"
	"strings"
)

// DestructiveQueryInfo contains analysis results for potentially dangerous SQL queries.
// It indicates whether a query could cause data loss and provides context about the risk.
type DestructiveQueryInfo struct {
	IsDestructive bool
	QueryType     string
	Warning       string
}

var (
	lineCommentRegex   = regexp.MustCompile(`--.*$`)
	blockCommentRegex  = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	whereClauseRegex   = regexp.MustCompile(`(?i)\bWHERE\b`)
	singleQuoteRegex   = regexp.MustCompile(`'(?:[^']|'')*'`)
	doubleQuoteRegex   = regexp.MustCompile(`"(?:[^"]|"")*"`)
	whitespaceRunRegex = regexp.MustCompile(`\s+`)
)

// AnalyzeQuerySafety examines a SQL statement and determines if it could cause data loss.
// It detects DROP, TRUNCATE, and DELETE/UPDATE statements without WHERE clauses.
// String literals are stripped before analysis to prevent false negatives from
// WHERE keywords appearing inside quoted values.
func AnalyzeQuerySafety(sql string) DestructiveQueryInfo {
	cleaned := cleanSQL(sql)
	upper := strings.ToUpper(cleaned)

	if strings.HasPrefix(upper, "DROP ") {
		return DestructiveQueryInfo{
			IsDestructive: true,
			QueryType:     "DROP",
			Warning:       "This will permanently remove database objects",
		}
	}

	if strings.HasPrefix(upper, "TRUNCATE ") {
		return DestructiveQueryInfo{
			IsDestructive: true,
			QueryType:     "TRUNCATE",
			Warning:       "This will delete all rows from the table",
		}
	}

	// Strip string literals before checking for WHERE clause to avoid
	// false negatives like UPDATE t SET col='WHERE' being marked safe
	strippedSQL := stripStringLiterals(cleaned)

	if strings.HasPrefix(upper, "DELETE ") {
		if !whereClauseRegex.MatchString(strippedSQL) {
			return DestructiveQueryInfo{
				IsDestructive: true,
				QueryType:     "DELETE",
				Warning:       "DELETE without WHERE clause will remove all rows",
			}
		}
	}

	if strings.HasPrefix(upper, "UPDATE ") {
		if !whereClauseRegex.MatchString(strippedSQL) {
			return DestructiveQueryInfo{
				IsDestructive: true,
				QueryType:     "UPDATE",
				Warning:       "UPDATE without WHERE clause will modify all rows",
			}
		}
	}

	return DestructiveQueryInfo{IsDestructive: false}
}

// stripStringLiterals removes single-quoted strings and double-quoted identifiers
// from SQL to prevent quoted content from interfering with keyword detection.
func stripStringLiterals(sql string) string {
	result := singleQuoteRegex.ReplaceAllString(sql, "''")
	result = doubleQuoteRegex.ReplaceAllString(result, "\"\"")
	return result
}

func cleanSQL(sql string) string {
	cleaned := blockCommentRegex.ReplaceAllString(sql, " ")

	lines := strings.Split(cleaned, "\n")
	var result []string
	for _, line := range lines {
		line = lineCommentRegex.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	joined := strings.Join(result, " ")
	// Normalize whitespace: collapse runs of spaces to single space
	return whitespaceRunRegex.ReplaceAllString(joined, " ")
}
