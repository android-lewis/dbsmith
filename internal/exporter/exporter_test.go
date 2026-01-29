package exporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/android-lewis/dbsmith/internal/models"
)

func TestCSVExporter_Export(t *testing.T) {
	tests := []struct {
		name     string
		result   *models.QueryResult
		want     string
		wantErr  bool
		errMatch string
	}{
		{
			name: "basic export",
			result: &models.QueryResult{
				Columns: []string{"id", "name", "age"},
				Rows: [][]interface{}{
					{1, "Alice", 30},
					{2, "Bob", 25},
				},
			},
			want: "id,name,age\n1,Alice,30\n2,Bob,25\n",
		},
		{
			name: "with null values",
			result: &models.QueryResult{
				Columns: []string{"id", "value"},
				Rows: [][]interface{}{
					{1, nil},
					{2, "test"},
				},
			},
			want: "id,value\n1,\n2,test\n", // nil is exported as empty string
		},
		{
			name: "with special characters",
			result: &models.QueryResult{
				Columns: []string{"id", "description"},
				Rows: [][]interface{}{
					{1, "hello, world"},
					{2, "line1\nline2"},
					{3, `with "quotes"`},
				},
			},
			want: "id,description\n1,\"hello, world\"\n2,\"line1\nline2\"\n3,\"with \"\"quotes\"\"\"\n",
		},
		{
			name: "empty rows",
			result: &models.QueryResult{
				Columns: []string{"id"},
				Rows:    [][]interface{}{},
			},
			wantErr:  true,
			errMatch: "no rows to export",
		},
		{
			name: "single row",
			result: &models.QueryResult{
				Columns: []string{"only_col"},
				Rows: [][]interface{}{
					{"only_value"},
				},
			},
			want: "only_col\nonly_value\n",
		},
		{
			name: "numeric types",
			result: &models.QueryResult{
				Columns: []string{"int", "float", "bool"},
				Rows: [][]interface{}{
					{int64(42), float64(3.14), true},
					{int64(-1), float64(0.0), false},
				},
			},
			want: "int,float,bool\n42,3.14,true\n-1,0,false\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exporter := NewCSVExporter(&buf)
			err := exporter.Export(tt.result)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMatch)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("CSV mismatch:\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestJSONExporter_Export(t *testing.T) {
	tests := []struct {
		name    string
		result  *models.QueryResult
		check   func(t *testing.T, output map[string]interface{})
		wantErr bool
	}{
		{
			name: "basic export",
			result: &models.QueryResult{
				Columns:     []string{"id", "name"},
				ExecutionMs: 42,
				Rows: [][]interface{}{
					{1, "Alice"},
					{2, "Bob"},
				},
			},
			check: func(t *testing.T, output map[string]interface{}) {
				if output["row_count"].(float64) != 2 {
					t.Errorf("expected row_count 2, got %v", output["row_count"])
				}
				if output["execution_ms"].(float64) != 42 {
					t.Errorf("expected execution_ms 42, got %v", output["execution_ms"])
				}

				columns := output["columns"].([]interface{})
				if len(columns) != 2 {
					t.Errorf("expected 2 columns, got %d", len(columns))
				}

				rows := output["rows"].([]interface{})
				if len(rows) != 2 {
					t.Errorf("expected 2 rows, got %d", len(rows))
				}

				firstRow := rows[0].(map[string]interface{})
				if firstRow["id"].(float64) != 1 {
					t.Errorf("expected first row id=1, got %v", firstRow["id"])
				}
				if firstRow["name"].(string) != "Alice" {
					t.Errorf("expected first row name=Alice, got %v", firstRow["name"])
				}
			},
		},
		{
			name: "empty rows",
			result: &models.QueryResult{
				Columns:     []string{"id"},
				ExecutionMs: 0,
				Rows:        [][]interface{}{},
			},
			check: func(t *testing.T, output map[string]interface{}) {
				if output["row_count"].(float64) != 0 {
					t.Errorf("expected row_count 0, got %v", output["row_count"])
				}
				rows := output["rows"].([]interface{})
				if len(rows) != 0 {
					t.Errorf("expected 0 rows, got %d", len(rows))
				}
			},
		},
		{
			name: "with null values",
			result: &models.QueryResult{
				Columns: []string{"id", "nullable"},
				Rows: [][]interface{}{
					{1, nil},
				},
			},
			check: func(t *testing.T, output map[string]interface{}) {
				rows := output["rows"].([]interface{})
				firstRow := rows[0].(map[string]interface{})
				if firstRow["nullable"] != nil {
					t.Errorf("expected null value, got %v", firstRow["nullable"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exporter := NewJSONExporter(&buf)
			err := exporter.Export(tt.result)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var output map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
				t.Fatalf("failed to parse JSON output: %v", err)
			}

			if tt.check != nil {
				tt.check(t, output)
			}
		})
	}
}

func TestExportToFormat(t *testing.T) {
	result := &models.QueryResult{
		Columns: []string{"id"},
		Rows: [][]interface{}{
			{1},
		},
	}

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"csv lowercase", "csv", false},
		{"csv uppercase", "CSV", false},
		{"json lowercase", "json", false},
		{"json uppercase", "JSON", false},
		{"json mixed case", "Json", false},
		{"unsupported format", "xml", true},
		{"empty format", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ExportToFormat(&buf, result, tt.format)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if buf.Len() == 0 {
					t.Error("expected non-empty output")
				}
			}
		})
	}
}

func TestRowsToJSON(t *testing.T) {
	columns := []string{"a", "b", "c"}
	rows := [][]interface{}{
		{1, "two", 3.0},
		{4, "five", nil},
	}

	result := rowsToJSON(columns, rows)

	if len(result) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result))
	}

	// Check first row
	if result[0]["a"] != 1 {
		t.Errorf("expected a=1, got %v", result[0]["a"])
	}
	if result[0]["b"] != "two" {
		t.Errorf("expected b=two, got %v", result[0]["b"])
	}

	// Check second row with nil
	if result[1]["c"] != nil {
		t.Errorf("expected c=nil, got %v", result[1]["c"])
	}
}

func TestRowsToJSON_MismatchedColumns(t *testing.T) {
	columns := []string{"a", "b", "c", "d"} // 4 columns
	rows := [][]interface{}{
		{1, 2}, // Only 2 values
	}

	result := rowsToJSON(columns, rows)

	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}

	// Should have values for a and b only
	if result[0]["a"] != 1 {
		t.Errorf("expected a=1, got %v", result[0]["a"])
	}
	if result[0]["b"] != 2 {
		t.Errorf("expected b=2, got %v", result[0]["b"])
	}
	if _, exists := result[0]["c"]; exists {
		t.Error("c should not exist")
	}
}
