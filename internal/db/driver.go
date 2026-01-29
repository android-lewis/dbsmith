package db

import (
	"context"

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

type BaseDriver struct {
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
