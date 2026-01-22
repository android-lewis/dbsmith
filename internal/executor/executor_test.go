package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/android-lewis/dbsmith/internal/constants"
	"github.com/android-lewis/dbsmith/internal/db"
	"github.com/android-lewis/dbsmith/internal/models"
)

func TestNewQueryExecutor(t *testing.T) {
	driver := db.NewMockDriver()
	qe := NewQueryExecutor(driver)

	if qe == nil {
		t.Error("Expected executor to be created")
		return
	}

	if qe.maxResults != 10000 {
		t.Errorf("Expected maxResults=10000, got %d", qe.maxResults)
	}

	if qe.timeout != 30*time.Second {
		t.Errorf("Expected timeout=30s, got %v", qe.timeout)
	}
}

func TestExecuteQuery(t *testing.T) {
	driver := db.NewMockDriver()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	_ = driver.Connect(context.Background(), conn, nil)

	qe := NewQueryExecutor(driver)

	result, err := qe.ExecuteQuery(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	if result.ExecutionMs < 0 {
		t.Errorf("Expected ExecutionMs >= 0, got %d", result.ExecutionMs)
	}
}

func TestExecuteQueryDisconnected(t *testing.T) {
	driver := db.NewMockDriver()
	qe := NewQueryExecutor(driver)

	_, err := qe.ExecuteQuery(context.Background(), "SELECT * FROM users")
	if err == nil {
		t.Error("Expected error when disconnected")
	}
}

func TestExecuteNonQuery(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	_ = driver.Connect(context.Background(), conn, nil)

	qe := NewQueryExecutor(driver)

	rowsAffected, err := qe.ExecuteNonQuery(context.Background(), "DELETE FROM users WHERE id = 1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", rowsAffected)
	}
}

func TestExecuteTransaction(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	_ = driver.Connect(context.Background(), conn, nil)

	qe := NewQueryExecutor(driver)

	queries := []string{
		"INSERT INTO users (name) VALUES ('Alice')",
		"INSERT INTO users (name) VALUES ('Bob')",
	}

	err := qe.ExecuteTransaction(context.Background(), queries)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestPing(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	err := driver.Connect(context.Background(), conn, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	qe := NewQueryExecutor(driver)

	err = qe.Ping(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestSetMaxResults(t *testing.T) {
	driver := db.NewMockDriver()
	qe := NewQueryExecutor(driver)

	qe.SetMaxResults(5000)
	if qe.maxResults != 5000 {
		t.Errorf("Expected maxResults=5000, got %d", qe.maxResults)
	}
}

func TestSetTimeout(t *testing.T) {
	driver := db.NewMockDriver()
	qe := NewQueryExecutor(driver)

	qe.SetTimeout(60 * time.Second)
	if qe.timeout != 60*time.Second {
		t.Errorf("Expected timeout=60s, got %v", qe.timeout)
	}
}

func TestIsConnected(t *testing.T) {
	driver := db.NewMockDriver()
	qe := NewQueryExecutor(driver)

	if qe.IsConnected() {
		t.Error("Expected driver to be disconnected initially")
	}

	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	err := driver.Connect(context.Background(), conn, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !qe.IsConnected() {
		t.Error("Expected driver to be connected")
	}
}

func TestClose(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	_ = driver.Connect(context.Background(), conn, nil)

	qe := NewQueryExecutor(driver)

	if !qe.IsConnected() {
		t.Error("Expected driver to be connected initially")
	}

	err := qe.Close(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if qe.IsConnected() {
		t.Error("Expected driver to be disconnected after close")
	}
}

func TestExecuteQueryContextCancellation(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	err := driver.Connect(context.Background(), conn, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	qe := NewQueryExecutor(driver)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = qe.ExecuteQuery(ctx, "SELECT * FROM users")
	if err == nil {
		t.Fatal("Expected error when context is cancelled")
	}

	if err != constants.ErrQueryCancelled {
		t.Errorf("Expected ErrQueryCancelled, got %v", err)
	}
}

func TestExecuteQueryContextTimeout(t *testing.T) {
	driver := db.NewMockDriver()
	driver.SetQueryDelay("pg_sleep", 1*time.Second)
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	err := driver.Connect(context.Background(), conn, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	qe := NewQueryExecutor(driver)
	qe.SetTimeout(10 * time.Millisecond)

	ctx := context.Background()
	_, err = qe.ExecuteQuery(ctx, "SELECT pg_sleep(10)")
	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if !errors.Is(err, constants.ErrQueryTimeout) {
		t.Errorf("Expected ErrQueryTimeout, got %v", err)
	}
}

func TestExecuteNonQueryContextTimeout(t *testing.T) {
	driver := db.NewMockDriver()
	driver.SetQueryDelay("pg_sleep", 1*time.Second)
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	err := driver.Connect(context.Background(), conn, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	qe := NewQueryExecutor(driver)
	qe.SetTimeout(10 * time.Millisecond)

	ctx := context.Background()
	_, err = qe.ExecuteNonQuery(ctx, "DELETE FROM users WHERE pg_sleep(10) IS NOT NULL")
	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if !errors.Is(err, constants.ErrQueryTimeout) {
		t.Errorf("Expected ErrQueryTimeout, got %v", err)
	}
}

func TestExplain(t *testing.T) {
	driver := db.NewMockDriver()
	conn := &models.Connection{Name: "test", Type: models.PostgresType}
	err := driver.Connect(context.Background(), conn, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	qe := NewQueryExecutor(driver)
	ctx := context.Background()

	result, err := qe.GetQueryExecutionPlan(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.RowCount == 0 {
		t.Error("Expected explain to return rows")
	}
}
