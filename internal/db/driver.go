package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/android-lewis/dbsmith/internal/logging"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/secrets"
)

type Driver interface {
	Connect(ctx context.Context, conn *models.Connection, secretsMgr secrets.Manager) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	Ping(ctx context.Context) error
	ExecuteQuery(ctx context.Context, sql string, args ...any) (*models.QueryResult, error)
	ExecuteNonQuery(ctx context.Context, sql string, args ...any) (int64, error)
	GetSchemas(ctx context.Context) ([]models.Schema, error)
	GetTables(ctx context.Context, schema models.Schema) ([]models.Table, error)
	GetTableColumns(ctx context.Context, tableName string) (*models.TableColumns, error)
	GetTableData(ctx context.Context, tableName string, limit int, offset int) (*models.QueryResult, error)
	GetTableIndexes(ctx context.Context, table string) ([]models.Index, error)
	ExecuteTransaction(ctx context.Context, queries []string) error
	GetVersion(ctx context.Context) (string, error)
	GetServerInfo(ctx context.Context) (*models.ServerInfo, error)
	GetQueryExecutionPlan(ctx context.Context, sql string) (*models.QueryResult, error)
	GetConnection() *models.Connection
}

// BaseDriver provides common functionality for all database drivers.
// Embed this in concrete driver implementations to inherit shared behavior.
type BaseDriver struct {
	db         *sql.DB
	connection *models.Connection
	connected  bool
}

func (bd *BaseDriver) GetConnection() *models.Connection {
	return bd.connection
}

func (bd *BaseDriver) IsConnected() bool {
	return bd.connected
}

func (bd *BaseDriver) setConnected(connected bool) {
	bd.connected = connected
}

func (bd *BaseDriver) setConnection(conn *models.Connection) {
	bd.connection = conn
}

func (bd *BaseDriver) validateConnection(conn *models.Connection, expectedType models.ConnectionType) error {
	if conn == nil {
		return ErrInvalidConnection
	}
	if conn.Type != expectedType {
		return ErrInvalidConnection
	}
	return nil
}

// DB returns the underlying *sql.DB for driver-specific operations.
func (bd *BaseDriver) BaseDb() *sql.DB {
	return bd.db
}

// ConnectWithDSN opens a database connection using the provided driver name and DSN.
// This is a helper for concrete driver Connect implementations.
func (bd *BaseDriver) ConnectWithDSN(ctx context.Context, driverName, dsn string, conn *models.Connection) error {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	if err := db.PingContext(ctx); err != nil {
		// Close error is secondary to ping failure - log but don't change the returned error
		if closeErr := db.Close(); closeErr != nil {
			logging.Debug().Err(closeErr).Msg("failed to close db after ping failure")
		}
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	bd.db = db
	bd.setConnection(conn)
	bd.setConnected(true)
	return nil
}

// Disconnect closes the database connection.
func (bd *BaseDriver) Disconnect(ctx context.Context) error {
	if bd.db != nil {
		if err := bd.db.Close(); err != nil {
			return fmt.Errorf("failed to disconnect: %w", err)
		}
	}
	bd.setConnected(false)
	return nil
}

// Ping verifies the database connection is still alive.
func (bd *BaseDriver) Ping(ctx context.Context) error {
	if !bd.IsConnected() || bd.db == nil {
		return ErrNotConnected
	}
	return bd.db.PingContext(ctx)
}

// ExecuteQuery runs a query and returns the result set.
func (bd *BaseDriver) ExecuteQuery(ctx context.Context, query string, args ...any) (*models.QueryResult, error) {
	if !bd.IsConnected() || bd.db == nil {
		return nil, ErrNotConnected
	}

	rows, err := bd.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer closeRows(rows)

	return scanRowsToResult(rows)
}

// ExecuteNonQuery runs a statement that doesn't return rows (INSERT, UPDATE, DELETE).
func (bd *BaseDriver) ExecuteNonQuery(ctx context.Context, query string, args ...any) (int64, error) {
	if !bd.IsConnected() || bd.db == nil {
		return 0, ErrNotConnected
	}

	result, err := bd.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}

	return result.RowsAffected()
}

// ExecuteTransaction runs multiple queries in a transaction.
func (bd *BaseDriver) ExecuteTransaction(ctx context.Context, queries []string) error {
	if !bd.IsConnected() || bd.db == nil {
		return ErrNotConnected
	}
	return executeTransaction(ctx, bd.db, queries)
}
