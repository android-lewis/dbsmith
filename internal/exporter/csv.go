package exporter

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/util"
)

type CSVExporter struct {
	writer *csv.Writer
}

func NewCSVExporter(w io.Writer) *CSVExporter {
	return &CSVExporter{
		writer: csv.NewWriter(w),
	}
}

func (e *CSVExporter) Export(result *models.QueryResult) error {
	if len(result.Rows) == 0 {
		return fmt.Errorf("no rows to export")
	}

	headers := result.Columns
	if err := e.writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, row := range result.Rows {
		record := make([]string, len(headers))
		for i, val := range row {
			record[i] = util.FormatValue(val)
		}
		if err := e.writer.Write(record); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	e.writer.Flush()
	return e.writer.Error()
}
