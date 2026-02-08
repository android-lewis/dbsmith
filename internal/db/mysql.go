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
}

func NewMySQLDriver() *MySQLDriver {
	return &MySQLDriver{}
}

func validateMySQLIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("%w: cannot be empty", ErrInvalidIdentifier)
	}

	pattern := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`)
	if !pattern.MatchString(identifier) {
		return fmt.Errorf("%w: contains unsafe characters", ErrInvalidIdentifier)
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

	return d.ConnectWithDSN(ctx, "mysql", dsn, conn)
}

func (d *MySQLDriver) GetSchemas(ctx context.Context) ([]models.Schema, error) {
	if !d.IsConnected() || d.BaseDb() == nil {
		return nil, ErrNotConnected
	}

	query := `SELECT schema_name, DEFAULT_CHARACTER_SET_NAME FROM information_schema.schemata;`

	rows, err := d.BaseDb().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer closeRows(rows)

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
	if !d.IsConnected() || d.BaseDb() == nil {
		return nil, ErrNotConnected
	}

	query := "SELECT TABLE_NAME, TABLE_SCHEMA, TABLE_TYPE FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME"

	rows, err := d.BaseDb().QueryContext(ctx, query, schema.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer closeRows(rows)

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

func (d *MySQLDriver) GetTableColumns(ctx context.Context, schemaName, tableName string) (*models.TableColumns, error) {
	if !d.IsConnected() || d.BaseDb() == nil {
		return nil, ErrNotConnected
	}

	tableName = strings.TrimSpace(tableName)
	schemaName = strings.TrimSpace(schemaName)

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
		WHERE TABLE_NAME = ? AND TABLE_SCHEMA = ?
		ORDER BY ORDINAL_POSITION
	`

	// If schemaName is empty, use current database
	if schemaName == "" {
		if err := d.BaseDb().QueryRowContext(ctx, "SELECT DATABASE()").Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("failed to get current database: %w", err)
		}
	}

	rows, err := d.BaseDb().QueryContext(ctx, columnQuery, tableName, schemaName)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

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
	ddlQuery := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", schemaName, tableName)
	var createTableDDL string
	if err := d.BaseDb().QueryRowContext(ctx, ddlQuery).Scan(&tableName, &createTableDDL); err != nil {
		logging.Warn().
			Err(err).
			Str("table", tableName).
			Str("schema", schemaName).
			Msg("Failed to retrieve DDL for table, continuing with empty DDL")
	}
	ddl = createTableDDL

	return &models.TableColumns{
		Columns: columns,
		DDL:     ddl,
	}, nil
}

func (d *MySQLDriver) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*models.QueryResult, error) {
	if !d.IsConnected() || d.BaseDb() == nil {
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
	if !d.IsConnected() || d.BaseDb() == nil {
		return nil, ErrNotConnected
	}

	query := "SHOW DATABASES"

	rows, err := d.BaseDb().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

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

func (d *MySQLDriver) GetVersion(ctx context.Context) (string, error) {
	if !d.IsConnected() || d.BaseDb() == nil {
		return "", ErrNotConnected
	}

	var version string
	err := d.BaseDb().QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	return version, err
}

func (d *MySQLDriver) GetServerInfo(ctx context.Context) (*models.ServerInfo, error) {
	if !d.IsConnected() || d.BaseDb() == nil {
		return nil, ErrNotConnected
	}

	info := &models.ServerInfo{
		ServerType:     "MySQL",
		AdditionalInfo: make(map[string]string),
	}

	// Get version
	if version, err := d.GetVersion(ctx); err == nil {
		info.Version = version
	}

	// Get current database and user (informational - errors are acceptable)
	if err := d.BaseDb().QueryRowContext(ctx, "SELECT DATABASE()").Scan(&info.CurrentDatabase); err != nil {
		logging.Debug().Err(err).Msg("could not fetch current database")
	}
	if err := d.BaseDb().QueryRowContext(ctx, "SELECT USER()").Scan(&info.CurrentUser); err != nil {
		logging.Debug().Err(err).Msg("could not fetch current user")
	}

	// Get uptime (in seconds, convert to duration string)
	var uptimeSeconds int64
	if err := d.BaseDb().QueryRowContext(ctx, "SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Uptime'").Scan(&uptimeSeconds); err == nil {
		hours := uptimeSeconds / 3600
		minutes := (uptimeSeconds % 3600) / 60
		seconds := uptimeSeconds % 60
		info.Uptime = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}

	// Get connection counts (informational - errors are acceptable)
	if err := d.BaseDb().QueryRowContext(ctx, "SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Threads_connected'").Scan(&info.ConnectionCount); err != nil {
		logging.Debug().Err(err).Msg("could not fetch connection count")
	}

	// Get max connections (informational - errors are acceptable)
	if err := d.BaseDb().QueryRowContext(ctx, "SELECT @@max_connections").Scan(&info.MaxConnections); err != nil {
		logging.Debug().Err(err).Msg("could not fetch max connections")
	}

	// Get database size
	if info.CurrentDatabase != "" {
		var size float64
		if err := d.BaseDb().QueryRowContext(ctx, `
			SELECT SUM(data_length + index_length) / 1024 / 1024 
			FROM information_schema.tables 
			WHERE table_schema = ?`, info.CurrentDatabase).Scan(&size); err == nil {
			info.DatabaseSize = fmt.Sprintf("%.2f MB", size)
		}
	}

	// Additional MySQL-specific info
	var charset string
	if err := d.BaseDb().QueryRowContext(ctx, "SELECT @@character_set_database").Scan(&charset); err == nil {
		info.AdditionalInfo["Charset"] = charset
	}

	var collation string
	if err := d.BaseDb().QueryRowContext(ctx, "SELECT @@collation_database").Scan(&collation); err == nil {
		info.AdditionalInfo["Collation"] = collation
	}

	return info, nil
}

func (d *MySQLDriver) GetQueryExecutionPlan(ctx context.Context, sql string) (*models.QueryResult, error) {
	if !d.IsConnected() || d.BaseDb() == nil {
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
