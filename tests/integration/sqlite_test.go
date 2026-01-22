//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"
)

// =============================================================================
// Connection Tests
// =============================================================================

func TestSQLiteConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupSQLite(t)
	defer setup.Close()

	driver := setup.Driver()
	if driver == nil {
		t.Fatal("Expected driver to be initialized")
	}

	if setup.SecretsManager() == nil {
		t.Fatal("Expected secrets manager to be initialized")
	}

	conn := setup.Connection()
	if conn == nil || conn.Name != "test-sqlite" {
		t.Fatal("Expected valid connection metadata")
	}

	if setup.DBType() != "sqlite" {
		t.Errorf("Expected dbType 'sqlite', got '%s'", setup.DBType())
	}
}

// =============================================================================
// Basic Query Tests
// =============================================================================

func TestSQLiteQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupSQLite(t)
	defer setup.Close()

	result := setup.ExecuteQuery(t, "SELECT 1 as test_value")

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	if len(result.Rows) == 0 {
		t.Fatal("Expected at least one row in result")
	}

	if len(result.Columns) == 0 {
		t.Fatal("Expected at least one column in result")
	}
}

// =============================================================================
// Fixture Loading Tests
// =============================================================================

func TestSQLiteLoadFixture(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupSQLite(t)
	defer setup.Close()

	// Load the fixture
	setup.LoadFixtureForDBType(t)

	// Verify data was loaded
	result := setup.ExecuteQuery(t, "SELECT COUNT(*) as count FROM users")
	if len(result.Rows) == 0 {
		t.Fatal("Expected at least one row")
	}

	count := result.Rows[0][0]
	if count == int64(0) || count == 0 {
		t.Fatal("Expected users table to have data after loading fixture")
	}

	// Test cleanup
	setup.Cleanup(t)

	// Verify data was cleaned up
	result = setup.ExecuteQuery(t, "SELECT COUNT(*) as count FROM users")
	count = result.Rows[0][0]
	if count != int64(0) && count != 0 {
		t.Errorf("Expected users table to be empty after cleanup, got %v", count)
	}
}

// =============================================================================
// Query Cancellation Tests
// =============================================================================

func TestSQLiteQueryCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupSQLite(t)
	defer setup.Close()

	driver := setup.Driver()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := driver.ExecuteQuery(ctx, "SELECT 1")
	if err == nil {
		t.Fatal("Expected error when context is cancelled")
	}

	if ctx.Err() != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", ctx.Err())
	}
}

// =============================================================================
// Schema Operations Tests
// =============================================================================

func TestSQLiteSchemaOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupSQLite(t)
	defer setup.Close()

	// Load fixture first
	setup.LoadFixtureForDBType(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	driver := setup.Driver()

	// Test getting schemas (SQLite returns a single schema with empty name)
	schemas, err := driver.GetSchemas(ctx)
	if err != nil {
		t.Fatalf("Failed to get schemas: %v", err)
	}
	if len(schemas) == 0 {
		t.Fatal("Expected at least one schema")
	}

	// Use the first schema (SQLite has a single default schema)
	schema := schemas[0]

	// Test getting tables
	tables, err := driver.GetTables(ctx, schema)
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}
	if len(tables) == 0 {
		t.Fatal("Expected at least one table after loading fixture")
	}

	// Find users table
	var foundUsers bool
	for _, table := range tables {
		if table.Name == "users" {
			foundUsers = true
			break
		}
	}
	if !foundUsers {
		t.Fatal("Expected to find 'users' table")
	}

	// Test getting columns
	columns, err := driver.GetTableColumns(ctx, "users")
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}
	if len(columns.Columns) == 0 {
		t.Fatal("Expected at least one column in users table")
	}

	// Cleanup
	setup.Cleanup(t)
}

// =============================================================================
// CRUD Operations Tests
// =============================================================================

func TestSQLiteCRUDOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupSQLite(t)
	defer setup.Close()

	// Load fixture to create tables
	setup.LoadFixtureForDBType(t)
	// Clean up seed data so we start fresh
	setup.Cleanup(t)

	// INSERT
	setup.MustExecute(t, "INSERT INTO users (email, name, age) VALUES ('test@test.com', 'Test User', 30)")

	// SELECT
	result := setup.ExecuteQuery(t, "SELECT id, email, name, age FROM users WHERE email = 'test@test.com'")
	if len(result.Rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(result.Rows))
	}

	// UPDATE
	setup.MustExecute(t, "UPDATE users SET age = 31 WHERE email = 'test@test.com'")

	result = setup.ExecuteQuery(t, "SELECT age FROM users WHERE email = 'test@test.com'")
	if len(result.Rows) != 1 {
		t.Fatalf("Expected 1 row after update, got %d", len(result.Rows))
	}

	// DELETE
	setup.MustExecute(t, "DELETE FROM users WHERE email = 'test@test.com'")

	result = setup.ExecuteQuery(t, "SELECT COUNT(*) FROM users WHERE email = 'test@test.com'")
	count := result.Rows[0][0]
	if count != int64(0) && count != 0 {
		t.Errorf("Expected 0 rows after delete, got %v", count)
	}
}
