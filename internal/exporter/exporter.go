package exporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/android-lewis/dbsmith/internal/models"
)

type Exporter interface {
	Export(result *models.QueryResult) error
}

type ExportOptions struct {
	TableName string
	BatchSize int
}

func ExportToFormat(w io.Writer, result *models.QueryResult, format string, opts ...ExportOptions) error {
	switch strings.ToLower(format) {
	case "csv":
		exporter := NewCSVExporter(w)
		return exporter.Export(result)
	case "json":
		exporter := NewJSONExporter(w)
		return exporter.Export(result)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}
