package db

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/android-lewis/dbsmith/internal/logging"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/secrets"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLDriver struct {
	BaseDriver
	db *sql.DB
}

func NewMySQLDriver() *MySQLDriver {
	return &MySQLDriver{}
}

func validateMySQLIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	pattern := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`)
	if !pattern.MatchString(identifier) {
		return fmt.Errorf("invalid identifier: contains unsafe characters")
	}
	return nil
}

func (d *MySQLDriver) Connect(ctx context.Context, conn *models.Connection, secretsMgr secrets.Manager) error {
	if err := d.validateConnection(conn, models.MySQLType); err != nil {
		return err
	}

	dsn, err := d.buildConnectionString(conn, secretsMgr)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	d.db = db
	d.setConnection(conn)
	d.setConnected(true)

	return nil
}

func (d *MySQLDriver) Disconnect(ctx context.Context) error {
	if d.db != nil {
		if err := d.db.Close(); err != nil {
			return fmt.Errorf("failed to disconnect: %v", err)
		}
	}

	d.setConnected(false)
	return nil
}

func (d *MySQLDriver) Ping(ctx context.Context) error {
	if !d.IsConnected() || d.db == nil {
		return ErrNotConnected
	}

	return d.db.PingContext(ctx)
}

func (d *MySQLDriver) ExecuteQuery(ctx context.Context, sql string, args ...interface{}) (*models.QueryResult, error) {
	if !d.IsConnected() || d.db == nil {
		return nil, ErrNotConnected
	}

	rows, err := d.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer func() { _ = rows.Close() }()

	return rowsToResult(rows)
}

func (d *MySQLDriver) ExecuteNonQuery(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	if !d.IsConnected() || d.db == nil {
		return 0, ErrNotConnected
	}

	result, err := d.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}

	return result.RowsAffected()
}

func (d *MySQLDriver) GetSchemas(ctx context.Context) ([]models.Schema, error) {
	if !d.IsConnected() || d.db == nil {
		return nil, ErrNotConnected
	}

	query := `SELECT schema_name, DEFAULT_CHARACTER_SET_NAME FROM information_schema.schemata;`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer func() { _ = rows.Close() }()

	var schemas []models.Schema
	for rows.Next() {
		var schema models.Schema
		var charset string
		if err := rows.Scan(&schema.Name, &charset); err != nil {
			return nil, err
		}
		schema.Owner = "root"

		schemas = append(schemas, schema)
	}

	return schemas, rows.Err()
}

func (d *MySQLDriver) GetTables(ctx context.Context, schema models.Schema) ([]models.Table, error) {
	if !d.IsConnected() || d.db == nil {
		return nil, ErrNotConnected
	}

	query := "SELECT TABLE_NAME, TABLE_SCHEMA, TABLE_TYPE FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME"

	rows, err := d.db.QueryContext(ctx, query, schema.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer func() { _ = rows.Close() }()

	var tables []models.Table
	for rows.Next() {
		var table models.Table
		if err := rows.Scan(&table.Name, &table.Schema, &table.Type); err != nil {
			return nil, err
		}

		tables = append(tables, table)
	}

	return tables, rows.Err()
}

func (d *MySQLDriver) GetTableColumns(ctx context.Context, tableName string) (*models.TableColumns, error) {
	if !d.IsConnected() || d.db == nil {
		return nil, ErrNotConnected
	}

	tableName = strings.TrimSpace(tableName)

	if err := validateMySQLIdentifier(tableName); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	columnQuery := `
		SELECT 
			COLUMN_NAME,
			COLUMN_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_KEY
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := d.db.QueryContext(ctx, columnQuery, tableName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var columns []models.Column
	for rows.Next() {
		var col models.Column
		var nullable string
		var defaultVal sql.NullString
		var columnKey string
		if err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal, &columnKey); err != nil {
			return nil, err
		}
		col.Nullable = nullable == "YES"
		if defaultVal.Valid {
			col.Default = defaultVal.String
		}
		col.IsPrimaryKey = columnKey == "PRI"
		col.IsForeignKey = columnKey == "MUL"
		columns = append(columns, col)
	}

	if len(columns) == 0 {
		return nil, ErrTableNotFound
	}

	var ddl string
	ddlQuery := fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)
	var createTableDDL string
	if err := d.db.QueryRowContext(ctx, ddlQuery).Scan(&tableName, &createTableDDL); err != nil {
		logging.Warn().
			Err(err).
			Str("table", tableName).
			Msg("Failed to retrieve DDL for table, continuing with empty DDL")
	}
	ddl = createTableDDL

	return &models.TableColumns{
		Columns: columns,
		DDL:     ddl,
	}, nil
}

func (d *MySQLDriver) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*models.QueryResult, error) {
	if !d.IsConnected() || d.db == nil {
		return nil, ErrNotConnected
	}

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", tableName, limit, offset)
	return d.ExecuteQuery(ctx, query)
}

func (d *MySQLDriver) GetDatabases(ctx context.Context) ([]string, error) {
	if !d.IsConnected() || d.db == nil {
		return nil, ErrNotConnected
	}

	query := "SHOW DATABASES"

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var dbs []string
	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err != nil {
			return nil, err
		}
		dbs = append(dbs, db)
	}

	return dbs, rows.Err()
}

func (d *MySQLDriver) ExecuteTransaction(ctx context.Context, queries []string) error {
	if !d.IsConnected() || d.db == nil {
		return ErrNotConnected
	}

	return executeTransaction(ctx, d.db, queries)
}

func (d *MySQLDriver) GetVersion(ctx context.Context) (string, error) {
	if !d.IsConnected() || d.db == nil {
		return "", ErrNotConnected
	}

	var version string
	err := d.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	return version, err
}

func (d *MySQLDriver) GetQueryExecutionPlan(ctx context.Context, sql string) (*models.QueryResult, error) {
	if !d.IsConnected() || d.db == nil {
		return nil, ErrNotConnected
	}

	explainSQL := "EXPLAIN " + sql
	return d.ExecuteQuery(ctx, explainSQL)
}

func (d *MySQLDriver) buildConnectionString(conn *models.Connection, secretsMgr secrets.Manager) (string, error) {
	var userPass string

	if conn.Username != "" {
		userPass = conn.Username
	}

	if conn.SecretKeyID != "" && secretsMgr != nil {
		password, err := secretsMgr.RetrieveSecret(conn.SecretKeyID)
		if err != nil {
			logging.Warn().
				Err(err).
				Str("secretKeyID", conn.SecretKeyID).
				Msg("Failed to retrieve password from secrets manager, connection will proceed without stored password")
		} else if password != "" {
			userPass += ":" + password
		}
	}

	if userPass != "" {
		userPass += "@"
	}

	var host string
	if conn.Host != "" {
		if conn.Port > 0 {
			host = fmt.Sprintf("tcp(%s:%d)", conn.Host, conn.Port)
		} else {
			host = fmt.Sprintf("tcp(%s:3306)", conn.Host)
		}
	}

	dbName := conn.Database
	if dbName == "" {
		dbName = "mysql"
	}

	params := "parseTime=true&loc=Local"

	return fmt.Sprintf("%s%s/%s?%s", userPass, host, dbName, params), nil
}

func (d *MySQLDriver) GetTableIndexes(ctx context.Context, table string) ([]models.Index, error) {
	if err := validateMySQLIdentifier(table); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	query := fmt.Sprintf("SHOW INDEXES FROM `%s`", table)
	result, err := d.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	indexMap := make(map[string]*models.Index)
	for _, row := range result.Rows {
		if len(row) >= 11 {
			indexName, _ := row[2].(string)
			colName, _ := row[4].(string)
			nonUnique, _ := row[1].(int64)
			indexType, _ := row[10].(string)

			if _, exists := indexMap[indexName]; !exists {
				indexMap[indexName] = &models.Index{
					Name:      indexName,
					Type:      indexType,
					Columns:   []string{},
					IsUnique:  nonUnique == 0,
					IsPrimary: indexName == "PRIMARY",
				}
			}
			indexMap[indexName].Columns = append(indexMap[indexName].Columns, colName)
		}
	}

	var indexes []models.Index
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}
	return indexes, nil
}
