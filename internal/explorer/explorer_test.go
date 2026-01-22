package explorer

import (
	"context"
	"testing"

	"github.com/android-lewis/dbsmith/internal/db"
	"github.com/android-lewis/dbsmith/internal/models"
)

func TestGetTables(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	_ = driver.Connect(context.Background(), conn, nil)

	qe := NewExplorer(driver)
	schema := models.Schema{
		Name:  "public",
		Owner: "test",
	}
	tables, err := qe.GetTables(context.Background(), schema)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tables) == 0 {
		t.Error("Expected at least one table")
	}
}

func TestGetTableSchema(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	_ = driver.Connect(context.Background(), conn, nil)

	qe := NewExplorer(driver)

	schema, err := qe.GetTableColumns(context.Background(), "users")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if schema == nil {
		t.Error("Expected schema to not be nil")
	}
}

func TestGetTableData(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	_ = driver.Connect(context.Background(), conn, nil)

	qe := NewExplorer(driver)

	result, err := qe.GetTableData(context.Background(), "users", 100, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected result to not be nil")
	}
}

func TestGetTableDataRespectMaxResults(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	_ = driver.Connect(context.Background(), conn, nil)

	qe := NewExplorer(driver)
	qe.SetMaxResults(100)

	_, _ = qe.GetTableData(context.Background(), "users", 50000, 0)

	if qe.maxResults != 100 {
		t.Errorf("Expected maxResults=100, got %d", qe.maxResults)
	}
}
