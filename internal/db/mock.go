package db

import (
	"context"
	"strings"
	"time"

	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/secrets"
)

type MockDriver struct {
	BaseDriver
	shouldFail      bool
	queryResults    map[string]*models.QueryResult
	queryRowResults map[string][]interface{}
	queryDelays     map[string]time.Duration
}

func NewMockDriver() *MockDriver {
	return &MockDriver{
		queryResults:    make(map[string]*models.QueryResult),
		queryRowResults: make(map[string][]interface{}),
		queryDelays:     make(map[string]time.Duration),
	}
}

func (md *MockDriver) SetQueryDelay(queryPattern string, delay time.Duration) {
	md.queryDelays[queryPattern] = delay
}

func (md *MockDriver) getDelay(sql string) time.Duration {
	for pattern, delay := range md.queryDelays {
		if strings.Contains(sql, pattern) {
			return delay
		}
	}
	return 0
}

func (md *MockDriver) waitWithContext(ctx context.Context, delay time.Duration) error {
	if delay == 0 {
		return nil
	}
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (md *MockDriver) AddQueryResult(query string, rows [][]interface{}, err error) {
	md.queryResults[query] = &models.QueryResult{
		Rows:     rows,
		Columns:  []string{},
		RowCount: int64(len(rows)),
	}
}

func (md *MockDriver) AddQueryRowResult(query string, row []interface{}, err error) {
	md.queryRowResults[query] = row
}

func (md *MockDriver) Connect(ctx context.Context, conn *models.Connection, secretsMgr secrets.Manager) error {
	if md.shouldFail {
		return ErrConnectionFailed
	}
	md.setConnection(conn)
	md.setConnected(true)
	return nil
}

func (md *MockDriver) Disconnect(ctx context.Context) error {
	md.setConnected(false)
	return nil
}

func (md *MockDriver) Ping(ctx context.Context) error {
	if !md.IsConnected() {
		return ErrNotConnected
	}
	return nil
}

func (md *MockDriver) ExecuteQuery(ctx context.Context, sql string, args ...interface{}) (*models.QueryResult, error) {
	if !md.IsConnected() {
		return nil, ErrNotConnected
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if err := md.waitWithContext(ctx, md.getDelay(sql)); err != nil {
		return nil, err
	}

	if result, ok := md.queryResults[sql]; ok {
		return result, nil
	}

	if row, ok := md.queryRowResults[sql]; ok {
		return &models.QueryResult{
			Columns:     []string{"result"},
			ColumnTypes: []string{""},
			Rows:        [][]interface{}{row},
			RowCount:    1,
		}, nil
	}

	return &models.QueryResult{
		Columns:     []string{"id", "name"},
		ColumnTypes: []string{"INTEGER", "TEXT"},
		Rows: [][]interface{}{
			{1, "test"},
		},
		RowCount: 1,
	}, nil
}

func (md *MockDriver) ExecuteQueryWithParameters(ctx context.Context, query *models.SavedQuery, params map[string]interface{}) (*models.QueryResult, error) {
	if !md.IsConnected() {
		return nil, ErrNotConnected
	}
	return &models.QueryResult{
		RowCount: 0,
	}, nil
}

func (md *MockDriver) ExecuteNonQuery(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	if !md.IsConnected() {
		return 0, ErrNotConnected
	}

	if err := md.waitWithContext(ctx, md.getDelay(sql)); err != nil {
		return 0, err
	}

	return 1, nil
}

func (md *MockDriver) GetSchemas(ctx context.Context) ([]models.Schema, error) {
	return []models.Schema{
		{
			Name:  "",
			Owner: "",
		},
	}, nil
}

func (md *MockDriver) GetTables(ctx context.Context, schema models.Schema) ([]models.Table, error) {
	if !md.IsConnected() {
		return nil, ErrNotConnected
	}
	return []models.Table{
		{
			Name:   "users",
			Schema: "public",
			Type:   "BASE TABLE",
		},
	}, nil
}

func (md *MockDriver) GetTableColumns(ctx context.Context, tableName string) (*models.TableColumns, error) {
	if !md.IsConnected() {
		return nil, ErrNotConnected
	}
	return &models.TableColumns{
		Columns: nil,
		DDL:     "CREATE TABLE " + tableName + " (...)",
	}, nil
}

func (md *MockDriver) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*models.QueryResult, error) {
	if !md.IsConnected() {
		return nil, ErrNotConnected
	}
	return &models.QueryResult{
		RowCount: 0,
	}, nil
}

func (md *MockDriver) GetDatabases(ctx context.Context) ([]string, error) {
	if !md.IsConnected() {
		return nil, ErrNotConnected
	}
	return []string{"postgres", "template0", "template1"}, nil
}

func (md *MockDriver) SelectDatabase(ctx context.Context, dbName string) error {
	if !md.IsConnected() {
		return ErrNotConnected
	}
	return nil
}

func (md *MockDriver) GetCurrentDatabase(ctx context.Context) (string, error) {
	if !md.IsConnected() {
		return "", ErrNotConnected
	}
	return "postgres", nil
}

func (md *MockDriver) ExecuteTransaction(ctx context.Context, queries []string) error {
	if !md.IsConnected() {
		return ErrNotConnected
	}
	return nil
}

func (md *MockDriver) GetVersion(ctx context.Context) (string, error) {
	if !md.IsConnected() {
		return "", ErrNotConnected
	}
	return "PostgreSQL 13.0", nil
}

func (md *MockDriver) GetServerInfo(ctx context.Context) (*models.ServerInfo, error) {
	if !md.IsConnected() {
		return nil, ErrNotConnected
	}
	return &models.ServerInfo{
		Version:         "PostgreSQL 13.0 (Mock)",
		ServerType:      "PostgreSQL",
		Uptime:          "00:00:00",
		CurrentDatabase: "mock_db",
		CurrentUser:     "mock_user",
		ConnectionCount: 1,
		MaxConnections:  100,
		DatabaseSize:    "0 MB",
		AdditionalInfo:  map[string]string{"Mode": "Mock"},
	}, nil
}

func (md *MockDriver) GetTableIndexes(ctx context.Context, table string) ([]models.Index, error) {
	if !md.IsConnected() {
		return nil, ErrNotConnected
	}
	return nil, nil
}

func (md *MockDriver) GetQueryExecutionPlan(ctx context.Context, sql string) (*models.QueryResult, error) {
	if !md.IsConnected() {
		return nil, ErrNotConnected
	}
	return &models.QueryResult{
		Columns:     []string{"QUERY PLAN"},
		ColumnTypes: []string{"TEXT"},
		Rows: [][]interface{}{
			{"Seq Scan on mock_table"},
		},
		RowCount: 1,
	}, nil
}
