package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/android-lewis/dbsmith/internal/models"
)

func rowsToResult(rows *sql.Rows) (*models.QueryResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	result := &models.QueryResult{
		Columns:     columns,
		ColumnTypes: make([]string, len(columnTypes)),
	}

	for i, ct := range columnTypes {
		result.ColumnTypes[i] = ct.DatabaseTypeName()
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		result.Rows = append(result.Rows, values)
		result.RowCount++
	}

	return result, rows.Err()
}

func executeTransaction(ctx context.Context, db *sql.DB, queries []string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, q := range queries {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}
	}

	return tx.Commit()
}
