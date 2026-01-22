//go:build integration
// +build integration

// Package integration provides test utilities for integration testing with real databases.
// It supports three database types: PostgreSQL, MySQL, and SQLite.
//
// PostgreSQL and MySQL tests use testcontainers for isolated, reproducible environments.
// SQLite tests use a temporary file-based database (no Docker required).
//
// Usage:
//
//	func TestPostgresFeature(t *testing.T) {
//	    setup := SetupPostgresContainer(t)
//	    defer setup.Close()
//	    // ... test code using setup.Driver(), setup.ExecuteQuery(), etc.
//	}
//
//	func TestSQLiteFeature(t *testing.T) {
//	    setup := SetupSQLite(t)
//	    defer setup.Close()
//	    // ... test code
//	}
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/android-lewis/dbsmith/internal/db"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/secrets"
)

// =============================================================================
// Core Types
// =============================================================================

// TestDatabaseSetup holds the configuration and connections for integration tests.
// It provides a common interface for all database types (PostgreSQL, MySQL, SQLite).
type TestDatabaseSetup struct {
	driver         db.Driver
	dsn            string
	dbType         string
	secretsManager secrets.Manager
	configDir      string
	connection     *models.Connection
}

// PostgresTestContainer wraps a testcontainers PostgreSQL instance with our driver.
type PostgresTestContainer struct {
	*TestDatabaseSetup
	container *postgres.PostgresContainer
}

// MySQLTestContainer wraps a testcontainers MySQL instance with our driver.
type MySQLTestContainer struct {
	*TestDatabaseSetup
	container *mysql.MySQLContainer
}

// =============================================================================
// Environment Detection
// =============================================================================

// IsCI returns true if running in a CI environment.
func IsCI() bool {
	ciVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"CIRCLECI",
		"TRAVIS",
		"JENKINS_URL",
		"BUILDKITE",
	}
	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}

// =============================================================================
// PostgreSQL Setup (Testcontainers)
// =============================================================================

// SetupPostgresContainer starts a PostgreSQL testcontainer with fixtures pre-loaded.
// The container is automatically cleaned up when the test completes.
func SetupPostgresContainer(t *testing.T) *PostgresTestContainer {
	t.Helper()

	ctx := context.Background()
	fixturePath := getFixturePath("postgres_schema.sql")

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts(fixturePath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("Warning: failed to terminate postgres container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	port, err := strconv.Atoi(mappedPort.Port())
	if err != nil {
		t.Fatalf("Failed to parse port: %v", err)
	}

	tmpDir := t.TempDir()
	secretsMgr, err := secrets.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secrets manager: %v", err)
	}

	passwordKey := "testcontainer-postgres-password"
	if err := secretsMgr.StoreSecret(passwordKey, "testpass"); err != nil {
		t.Fatalf("Failed to store password in secrets manager: %v", err)
	}

	conn := &models.Connection{
		Name:        "testcontainer-postgres",
		Type:        "postgres",
		Host:        host,
		Port:        port,
		Username:    "testuser",
		SecretKeyID: passwordKey,
		Database:    "testdb",
		SSL:         "disable",
	}

	driver := db.NewPostgresDriver()
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := driver.Connect(connectCtx, conn, secretsMgr); err != nil {
		t.Fatalf("Failed to connect to postgres container: %v", err)
	}

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%d/testdb?sslmode=disable", host, port)

	return &PostgresTestContainer{
		TestDatabaseSetup: &TestDatabaseSetup{
			driver:         driver,
			dsn:            dsn,
			dbType:         "postgres",
			secretsManager: secretsMgr,
			configDir:      tmpDir,
			connection:     conn,
		},
		container: container,
	}
}

// Close disconnects from the database. Container cleanup is handled by t.Cleanup.
func (ptc *PostgresTestContainer) Close() error {
	if ptc.driver != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return ptc.driver.Disconnect(ctx)
	}
	return nil
}

// Container returns the underlying testcontainer for advanced usage.
func (ptc *PostgresTestContainer) Container() *postgres.PostgresContainer {
	return ptc.container
}

// =============================================================================
// PostgreSQL TLS Setup (Testcontainers)
// =============================================================================

// PostgresTLSTestContainer wraps a TLS-enabled PostgreSQL testcontainer.
// Unlike PostgresTestContainer, this container requires SSL connections and
// rejects non-SSL connections.
type PostgresTLSTestContainer struct {
	*TestDatabaseSetup
	container  testcontainers.Container
	CACertPath string         // Path to CA cert for client verification
	certBundle *TLSCertBundle // Generated certificates
}

// SetupPostgresTLSContainer starts a PostgreSQL testcontainer with TLS enabled.
// The container requires SSL connections and rejects non-SSL connections.
// Certificates are generated at runtime using ECDSA P-256.
func SetupPostgresTLSContainer(t *testing.T) *PostgresTLSTestContainer {
	t.Helper()

	ctx := context.Background()

	// 1. Generate certificates
	certs := GenerateTLSCerts(t, "localhost", "127.0.0.1")

	// 2. Get config file paths
	sslConfPath := getFixturePath("tls/postgresql_ssl.conf")
	hbaConfPath := getFixturePath("tls/pg_hba_ssl.conf")
	initScriptPath := getFixturePath("postgres_schema.sql")

	// 3. Create startup script that fixes key permissions before starting postgres
	// PostgreSQL requires the server key to be owned by postgres user with 0600 permissions
	startupScript := `#!/bin/bash
set -e

# Copy certificates to postgres-owned location with correct permissions
cp /tmp/certs/server.crt /var/lib/postgresql/server.crt
cp /tmp/certs/server.key /var/lib/postgresql/server.key
chown postgres:postgres /var/lib/postgresql/server.crt /var/lib/postgresql/server.key
chmod 644 /var/lib/postgresql/server.crt
chmod 600 /var/lib/postgresql/server.key

# Execute the original entrypoint
exec docker-entrypoint.sh "$@"
`
	startupScriptPath := filepath.Join(certs.TempDir, "startup.sh")
	if err := os.WriteFile(startupScriptPath, []byte(startupScript), 0755); err != nil {
		t.Fatalf("Failed to write startup script: %v", err)
	}

	// 4. Create container with TLS config
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		Files: []testcontainers.ContainerFile{
			// Certificates (in temp location, will be copied by startup script)
			{HostFilePath: certs.ServerCertPath, ContainerFilePath: "/tmp/certs/server.crt", FileMode: 0644},
			{HostFilePath: certs.ServerKeyPath, ContainerFilePath: "/tmp/certs/server.key", FileMode: 0644},
			// Startup script
			{HostFilePath: startupScriptPath, ContainerFilePath: "/usr/local/bin/startup.sh", FileMode: 0755},
			// Config files
			{HostFilePath: sslConfPath, ContainerFilePath: "/etc/postgresql/postgresql.conf", FileMode: 0644},
			{HostFilePath: hbaConfPath, ContainerFilePath: "/etc/postgresql/pg_hba.conf", FileMode: 0644},
			// Init script
			{HostFilePath: initScriptPath, ContainerFilePath: "/docker-entrypoint-initdb.d/init.sql", FileMode: 0644},
		},
		Entrypoint: []string{"/usr/local/bin/startup.sh"},
		Cmd: []string{
			"postgres",
			"-c", "config_file=/etc/postgresql/postgresql.conf",
			"-c", "hba_file=/etc/postgresql/pg_hba.conf",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start TLS postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Warning: failed to terminate TLS postgres container: %v", err)
		}
	})

	// 4. Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	port, err := strconv.Atoi(mappedPort.Port())
	if err != nil {
		t.Fatalf("Failed to parse port: %v", err)
	}

	// 5. Setup secrets manager
	tmpDir := t.TempDir()
	secretsMgr, err := secrets.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secrets manager: %v", err)
	}

	passwordKey := "testcontainer-postgres-tls-password"
	if err := secretsMgr.StoreSecret(passwordKey, "testpass"); err != nil {
		t.Fatalf("Failed to store password: %v", err)
	}

	// 6. Create connection config (SSL mode will be set per-test)
	conn := &models.Connection{
		Name:        "testcontainer-postgres-tls",
		Type:        "postgres",
		Host:        host,
		Port:        port,
		Username:    "testuser",
		SecretKeyID: passwordKey,
		Database:    "testdb",
		SSL:         "require", // Default - tests will override
	}

	return &PostgresTLSTestContainer{
		TestDatabaseSetup: &TestDatabaseSetup{
			driver:         nil, // Driver created per-test with appropriate SSL mode
			dsn:            "",
			dbType:         "postgres",
			secretsManager: secretsMgr,
			configDir:      tmpDir,
			connection:     conn,
		},
		container:  container,
		CACertPath: certs.CACertPath,
		certBundle: certs,
	}
}

// ConnectWithSSLMode creates a new driver connection with the specified SSL mode.
// This is used by tests to verify different SSL modes work correctly.
func (ptc *PostgresTLSTestContainer) ConnectWithSSLMode(t *testing.T, sslMode string, caCertPath string) db.Driver {
	t.Helper()

	conn := *ptc.connection // Copy
	conn.SSL = sslMode
	conn.SSLCACertPath = caCertPath

	driver := db.NewPostgresDriver()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := driver.Connect(ctx, &conn, ptc.secretsManager)
	if err != nil {
		t.Fatalf("Failed to connect with SSL mode %s: %v", sslMode, err)
	}

	return driver
}

// TryConnectWithSSLMode attempts connection and returns error (for negative tests).
// Unlike ConnectWithSSLMode, this doesn't fail the test on error.
func (ptc *PostgresTLSTestContainer) TryConnectWithSSLMode(sslMode string, caCertPath string) error {
	conn := *ptc.connection
	conn.SSL = sslMode
	conn.SSLCACertPath = caCertPath

	driver := db.NewPostgresDriver()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return driver.Connect(ctx, &conn, ptc.secretsManager)
}

// Close for PostgresTLSTestContainer - container cleanup is handled by t.Cleanup()
func (ptc *PostgresTLSTestContainer) Close() error {
	// Container cleanup handled by t.Cleanup()
	return nil
}

// =============================================================================
// MySQL Setup (Testcontainers)
// =============================================================================

// SetupMySQLContainer starts a MySQL testcontainer with fixtures pre-loaded.
// The container is automatically cleaned up when the test completes.
func SetupMySQLContainer(t *testing.T) *MySQLTestContainer {
	t.Helper()

	ctx := context.Background()
	fixturePath := getFixturePath("mysql_schema.sql")

	container, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithDatabase("testdb"),
		mysql.WithUsername("testuser"),
		mysql.WithPassword("testpass"),
		mysql.WithScripts(fixturePath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  MySQL Community Server - GPL").
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start mysql container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("Warning: failed to terminate mysql container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "3306")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	port, err := strconv.Atoi(mappedPort.Port())
	if err != nil {
		t.Fatalf("Failed to parse port: %v", err)
	}

	tmpDir := t.TempDir()
	secretsMgr, err := secrets.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secrets manager: %v", err)
	}

	passwordKey := "testcontainer-mysql-password"
	if err := secretsMgr.StoreSecret(passwordKey, "testpass"); err != nil {
		t.Fatalf("Failed to store password in secrets manager: %v", err)
	}

	conn := &models.Connection{
		Name:        "testcontainer-mysql",
		Type:        "mysql",
		Host:        host,
		Port:        port,
		Username:    "testuser",
		SecretKeyID: passwordKey,
		Database:    "testdb",
	}

	driver := db.NewMySQLDriver()
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := driver.Connect(connectCtx, conn, secretsMgr); err != nil {
		t.Fatalf("Failed to connect to mysql container: %v", err)
	}

	dsn := fmt.Sprintf("testuser:testpass@tcp(%s:%d)/testdb?parseTime=true", host, port)

	return &MySQLTestContainer{
		TestDatabaseSetup: &TestDatabaseSetup{
			driver:         driver,
			dsn:            dsn,
			dbType:         "mysql",
			secretsManager: secretsMgr,
			configDir:      tmpDir,
			connection:     conn,
		},
		container: container,
	}
}

// Close disconnects from the database. Container cleanup is handled by t.Cleanup.
func (mtc *MySQLTestContainer) Close() error {
	if mtc.driver != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return mtc.driver.Disconnect(ctx)
	}
	return nil
}

// Container returns the underlying testcontainer for advanced usage.
func (mtc *MySQLTestContainer) Container() *mysql.MySQLContainer {
	return mtc.container
}

// =============================================================================
// SQLite Setup (No Docker Required)
// =============================================================================

// SetupSQLite creates a SQLite test database setup using a temporary file.
// This is always available and doesn't require Docker.
func SetupSQLite(t *testing.T) *TestDatabaseSetup {
	t.Helper()

	tmpDir := t.TempDir()
	secretsMgr, err := secrets.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secrets manager: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")

	driver := db.NewSQLiteDriver()
	conn := &models.Connection{
		Name:     "test-sqlite",
		Type:     "sqlite",
		Database: dbPath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := driver.Connect(ctx, conn, secretsMgr); err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}

	return &TestDatabaseSetup{
		driver:         driver,
		dsn:            dbPath,
		dbType:         "sqlite",
		secretsManager: secretsMgr,
		configDir:      tmpDir,
		connection:     conn,
	}
}

// =============================================================================
// TestDatabaseSetup Methods
// =============================================================================

// Close disconnects from the database and cleans up resources.
func (ts *TestDatabaseSetup) Close() error {
	if ts.driver != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return ts.driver.Disconnect(ctx)
	}
	return nil
}

// Driver returns the database driver.
func (ts *TestDatabaseSetup) Driver() db.Driver {
	return ts.driver
}

// DSN returns the data source name.
func (ts *TestDatabaseSetup) DSN() string {
	return ts.dsn
}

// SecretsManager returns the secrets manager.
func (ts *TestDatabaseSetup) SecretsManager() secrets.Manager {
	return ts.secretsManager
}

// ConfigDir returns the configuration directory.
func (ts *TestDatabaseSetup) ConfigDir() string {
	return ts.configDir
}

// Connection returns the connection configuration.
func (ts *TestDatabaseSetup) Connection() *models.Connection {
	return ts.connection
}

// DBType returns the database type (postgres, mysql, sqlite).
func (ts *TestDatabaseSetup) DBType() string {
	return ts.dbType
}

// =============================================================================
// Query Helpers
// =============================================================================

// ExecuteQuery executes a SQL query and returns the result.
func (ts *TestDatabaseSetup) ExecuteQuery(t *testing.T, query string) *models.QueryResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := ts.driver.ExecuteQuery(ctx, query)
	if err != nil {
		t.Fatalf("Query execution failed: %v", err)
	}

	return result
}

// MustExecute executes a SQL statement that doesn't return results.
func (ts *TestDatabaseSetup) MustExecute(t *testing.T, sql string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := ts.driver.ExecuteQuery(ctx, sql)
	if err != nil {
		t.Fatalf("SQL execution failed: %v", err)
	}
}

// =============================================================================
// Fixture Management
// =============================================================================

// LoadFixture loads a SQL fixture file into the database.
// It splits the file into individual statements and executes them separately
// to support databases like MySQL that don't allow multiple statements by default.
func (ts *TestDatabaseSetup) LoadFixture(t *testing.T, fixturePath string) {
	t.Helper()

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", fixturePath, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Split by semicolon and execute each statement individually
	statements := strings.Split(string(data), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		// Skip comment-only lines
		if strings.HasPrefix(stmt, "--") && !strings.Contains(stmt, "\n") {
			continue
		}
		_, err = ts.driver.ExecuteQuery(ctx, stmt)
		if err != nil {
			t.Fatalf("Failed to execute fixture %s: %v", fixturePath, err)
		}
	}
}

// LoadFixtureForDBType loads the appropriate fixture for the current database type.
func (ts *TestDatabaseSetup) LoadFixtureForDBType(t *testing.T) {
	t.Helper()

	fixturePath := getFixturePath(ts.dbType + "_schema.sql")
	ts.LoadFixture(t, fixturePath)
}

// Cleanup removes test data from the database (truncates tables).
func (ts *TestDatabaseSetup) Cleanup(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var queries []string
	switch ts.dbType {
	case "postgres":
		queries = []string{
			"TRUNCATE TABLE comments, posts, users RESTART IDENTITY CASCADE",
		}
	case "mysql":
		queries = []string{
			"SET FOREIGN_KEY_CHECKS = 0",
			"TRUNCATE TABLE comments",
			"TRUNCATE TABLE posts",
			"TRUNCATE TABLE users",
			"SET FOREIGN_KEY_CHECKS = 1",
		}
	case "sqlite":
		queries = []string{
			"DELETE FROM comments",
			"DELETE FROM posts",
			"DELETE FROM users",
			"DELETE FROM sqlite_sequence WHERE name IN ('comments', 'posts', 'users')",
		}
	}

	for _, query := range queries {
		_, err := ts.driver.ExecuteQuery(ctx, query)
		if err != nil {
			// Log but don't fail - cleanup is best effort
			t.Logf("Warning: cleanup query failed: %v", err)
		}
	}
}

// =============================================================================
// Credential Helpers
// =============================================================================

// StoreCredential stores a credential in the secrets manager.
func (ts *TestDatabaseSetup) StoreCredential(t *testing.T, keyID, value string) error {
	t.Helper()
	return ts.secretsManager.StoreSecret(keyID, value)
}

// RetrieveCredential retrieves a credential from the secrets manager.
func (ts *TestDatabaseSetup) RetrieveCredential(t *testing.T, keyID string) (string, error) {
	t.Helper()
	return ts.secretsManager.RetrieveSecret(keyID)
}

// =============================================================================
// Internal Helpers
// =============================================================================

// getFixturePath returns the absolute path to a fixture file.
func getFixturePath(fixtureName string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)
	return filepath.Join(currentDir, "..", "fixtures", fixtureName)
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
