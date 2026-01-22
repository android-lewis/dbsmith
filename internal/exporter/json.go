package exporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/android-lewis/dbsmith/internal/models"
)

type JSONExporter struct {
	writer io.Writer
}

func NewJSONExporter(w io.Writer) *JSONExporter {
	return &JSONExporter{writer: w}
}

func (e *JSONExporter) Export(result *models.QueryResult) error {
	data := map[string]interface{}{
		"columns":      result.Columns,
		"row_count":    len(result.Rows),
		"execution_ms": result.ExecutionMs,
		"rows":         rowsToJSON(result.Columns, result.Rows),
	}

	encoder := json.NewEncoder(e.writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func rowsToJSON(columns []string, rows [][]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		rowMap := make(map[string]interface{})
		for j, col := range columns {
			if j < len(row) {
				rowMap[col] = row[j]
			}
		}
		result[i] = rowMap
	}
	return result
}
