package util

import (
	"regexp"
	"strings"
)

// DestructiveQueryInfo contains details about a potentially destructive query
type DestructiveQueryInfo struct {
	IsDestructive bool
	QueryType     string // DELETE, DROP, TRUNCATE, UPDATE
	Warning       string
}

// whitespace and comment patterns
var (
	lineCommentRegex  = regexp.MustCompile(`--.*$`)
	blockCommentRegex = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	whereClauseRegex  = regexp.MustCompile(`(?i)\bWHERE\b`)
)

// AnalyzeQuerySafety checks if a SQL query is potentially destructive
func AnalyzeQuerySafety(sql string) DestructiveQueryInfo {
	// Remove comments and normalize whitespace
	cleaned := cleanSQL(sql)
	upper := strings.ToUpper(cleaned)

	// Check for DROP statements
	if strings.HasPrefix(upper, "DROP ") {
		return DestructiveQueryInfo{
			IsDestructive: true,
			QueryType:     "DROP",
			Warning:       "This will permanently remove database objects",
		}
	}

	// Check for TRUNCATE statements
	if strings.HasPrefix(upper, "TRUNCATE ") {
		return DestructiveQueryInfo{
			IsDestructive: true,
			QueryType:     "TRUNCATE",
			Warning:       "This will delete all rows from the table",
		}
	}

	// Check for DELETE without WHERE
	if strings.HasPrefix(upper, "DELETE ") {
		if !whereClauseRegex.MatchString(cleaned) {
			return DestructiveQueryInfo{
				IsDestructive: true,
				QueryType:     "DELETE",
				Warning:       "DELETE without WHERE clause will remove all rows",
			}
		}
	}

	// Check for UPDATE without WHERE
	if strings.HasPrefix(upper, "UPDATE ") {
		if !whereClauseRegex.MatchString(cleaned) {
			return DestructiveQueryInfo{
				IsDestructive: true,
				QueryType:     "UPDATE",
				Warning:       "UPDATE without WHERE clause will modify all rows",
			}
		}
	}

	return DestructiveQueryInfo{IsDestructive: false}
}

// cleanSQL removes comments and normalizes whitespace
func cleanSQL(sql string) string {
	// Remove block comments
	cleaned := blockCommentRegex.ReplaceAllString(sql, " ")

	// Process line by line to handle line comments
	lines := strings.Split(cleaned, "\n")
	var result []string
	for _, line := range lines {
		// Remove line comments
		line = lineCommentRegex.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, " ")
}
