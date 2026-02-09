package explorer

import (
	"context"
	"time"

	"github.com/android-lewis/dbsmith/internal/constants"
	"github.com/android-lewis/dbsmith/internal/db"
	"github.com/android-lewis/dbsmith/internal/models"
)

type Explorer struct {
	driver     db.Driver
	maxResults int
	timeout    time.Duration
}

func NewExplorer(driver db.Driver) *Explorer {
	return &Explorer{
		driver:     driver,
		maxResults: constants.DefaultMaxResults,
		timeout:    constants.DefaultTimeout,
	}
}

func (e *Explorer) GetTables(ctx context.Context, schema models.Schema) ([]models.Table, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	tables, err := e.driver.GetTables(ctx, schema)
	if err != nil {
		return nil, err
	}

	if len(tables) > e.maxResults {
		tables = tables[:e.maxResults]
	}

	return tables, nil
}

func (e *Explorer) GetTableColumns(ctx context.Context, schemaName, tableName string) (*models.TableColumns, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	return e.driver.GetTableColumns(ctx, schemaName, tableName)
}

func (e *Explorer) GetSchemas(ctx context.Context) ([]models.Schema, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	schemas, err := e.driver.GetSchemas(ctx)
	if err != nil {
		return nil, err
	}

	if len(schemas) > e.maxResults {
		schemas = schemas[:e.maxResults]
	}

	return schemas, nil
}

func (e *Explorer) GetTableIndexes(ctx context.Context, tableName string) ([]models.Index, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	return e.driver.GetTableIndexes(ctx, tableName)
}

func (e *Explorer) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*models.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	if limit > e.maxResults {
		limit = e.maxResults
	}

	return e.driver.GetTableData(ctx, tableName, limit, offset)
}

func (e *Explorer) SetMaxResults(max int) {
	e.maxResults = max
}

func (e *Explorer) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

func (e *Explorer) GetTimeout() time.Duration {
	return e.timeout
}
