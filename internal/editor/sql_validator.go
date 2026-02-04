package editor

import (
	"regexp"
	"strings"
)

type DestructiveQueryInfo struct {
	IsDestructive bool
	QueryType     string
	Warning       string
}

var (
	lineCommentRegex  = regexp.MustCompile(`--.*$`)
	blockCommentRegex = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	whereClauseRegex  = regexp.MustCompile(`(?i)\bWHERE\b`)
)

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

	if strings.HasPrefix(upper, "DELETE ") {
		if !whereClauseRegex.MatchString(cleaned) {
			return DestructiveQueryInfo{
				IsDestructive: true,
				QueryType:     "DELETE",
				Warning:       "DELETE without WHERE clause will remove all rows",
			}
		}
	}

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

	return strings.Join(result, " ")
}
