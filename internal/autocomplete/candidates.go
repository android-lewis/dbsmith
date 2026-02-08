package autocomplete

import (
	"strconv"
	"strings"
)

func buildCandidates(ctx completionContext, stmt statement, dbContext Context, dialect, qualifier, quote string) []Item {
	var items []Item

	if hasKind(ctx.kinds, KindColumn) {
		items = append(items, columnCandidates(stmt, dbContext, qualifier, quote)...)
	}
	if hasKind(ctx.kinds, KindTable) {
		items = append(items, tableCandidates(dbContext, quote)...)
	}
	if hasKind(ctx.kinds, KindKeyword) {
		items = append(items, keywordCandidates(dialect)...)
	}

	return dedupeItems(items)
}

func hasKind(kinds []ItemKind, kind ItemKind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

func keywordCandidates(dialect string) []Item {
	keywords := keywordsForDialect(dialect)
	items := make([]Item, 0, len(keywords))
	for _, keyword := range keywords {
		items = append(items, Item{
			Label:  keyword,
			Kind:   KindKeyword,
			Detail: "keyword",
		})
	}
	return items
}

func tableCandidates(ctx Context, quote string) []Item {
	items := make([]Item, 0, len(ctx.Tables))
	for _, table := range ctx.Tables {
		label := tableLabel(table, quote)
		detail := "table"
		if table.Schema != "" {
			detail = "table (" + table.Schema + ")"
		}
		items = append(items, Item{
			Label:  label,
			Kind:   KindTable,
			Detail: detail,
		})
	}
	return items
}

func columnCandidates(stmt statement, ctx Context, qualifier, quote string) []Item {
	columnsByTable := map[string][]Column{}
	for _, table := range ctx.Tables {
		columnsByTable[strings.ToUpper(table.Name)] = table.Columns
	}

	var targetTables []string
	if qualifier != "" {
		target := resolveQualifiedTable(qualifier, stmt)
		if target != "" {
			targetTables = append(targetTables, target)
		}
	}

	if len(targetTables) == 0 {
		for _, table := range stmt.Tables {
			targetTables = append(targetTables, table.Name)
		}
	}

	if len(targetTables) == 0 {
		for _, table := range ctx.Tables {
			targetTables = append(targetTables, table.Name)
		}
	}

	var items []Item
	for _, tableName := range targetTables {
		cols := columnsByTable[strings.ToUpper(tableName)]
		for _, col := range cols {
			label := col.Name
			if quote != "" {
				label = quote + label + quote
			}
			detail := "column"
			if tableName != "" {
				detail = "column (" + tableName + ")"
			}
			items = append(items, Item{
				Label:  label,
				Kind:   KindColumn,
				Detail: detail,
			})
		}
	}
	return items
}

func tableLabel(table Table, quote string) string {
	name := table.Name
	if table.Schema != "" {
		name = table.Schema + "." + table.Name
	}
	if quote != "" {
		return quote + name + quote
	}
	return name
}

func resolveQualifiedTable(qualifier string, stmt statement) string {
	if qualifier == "" {
		return ""
	}

	upper := strings.ToUpper(qualifier)
	if resolved, ok := stmt.Aliases[upper]; ok {
		return resolved
	}

	for _, table := range stmt.Tables {
		if strings.EqualFold(table.Name, qualifier) {
			return table.Name
		}
	}
	return ""
}

func filterCandidates(items []Item, prefix string) []Item {
	if prefix == "" {
		return items
	}

	filtered := make([]Item, 0, len(items))
	for _, item := range items {
		value := strings.Trim(item.Label, "`\"")
		upperValue := strings.ToUpper(value)
		upperPrefix := strings.ToUpper(prefix)
		matched := strings.HasPrefix(upperValue, upperPrefix)
		if !matched {
			if dot := strings.LastIndex(value, "."); dot >= 0 && dot+1 < len(value) {
				suffix := strings.ToUpper(value[dot+1:])
				matched = strings.HasPrefix(suffix, upperPrefix)
			}
		}
		if matched {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func dedupeItems(items []Item) []Item {
	seen := map[string]bool{}
	result := make([]Item, 0, len(items))
	for _, item := range items {
		key := strings.ToUpper(item.Label) + ":" + strings.ToUpper(item.Detail) + ":" + strconv.Itoa(int(item.Kind))
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}
	return result
}
