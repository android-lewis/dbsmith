package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/android-lewis/dbsmith/internal/logging"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/secrets"
	"github.com/lib/pq"
)

type PostgresDriver struct {
	BaseDriver
}

func NewPostgresDriver() *PostgresDriver {
	return &PostgresDriver{}
}

func (d *PostgresDriver) Connect(ctx context.Context, conn *models.Connection, secretsMgr secrets.Manager) error {
	if err := d.validateConnection(conn, models.PostgresType); err != nil {
		return err
	}

	dsn, err := d.buildConnectionString(conn, secretsMgr)
	if err != nil {
		return err
	}

	return d.ConnectWithDSN(ctx, "postgres", dsn, conn)
}

func (d *PostgresDriver) GetSchemas(ctx context.Context) ([]models.Schema, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	query := `SELECT schema_name, schema_owner FROM information_schema.schemata;`

	rows, err := d.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer func() { _ = rows.Close() }()

	var schemas []models.Schema
	for rows.Next() {
		var schema models.Schema
		if err := rows.Scan(&schema.Name, &schema.Owner); err != nil {
			return nil, err
		}

		schemas = append(schemas, schema)
	}

	return schemas, rows.Err()
}

func (d *PostgresDriver) GetTables(ctx context.Context, schema models.Schema) ([]models.Table, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	query := `
		SELECT table_name, table_schema, table_type
		FROM information_schema.tables
		WHERE table_schema = $1
		ORDER BY table_name
	`

	rows, err := d.DB().QueryContext(ctx, query, schema.Name)
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

func (d *PostgresDriver) GetTableColumns(ctx context.Context, tableName string) (*models.TableColumns, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	tableName = strings.TrimSpace(tableName)

	columnQuery := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
			CASE WHEN fk.column_name IS NOT NULL THEN true ELSE false END as is_foreign_key
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT kcu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu 
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_name = $1
		) pk ON c.column_name = pk.column_name
		LEFT JOIN (
			SELECT kcu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu 
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY'
				AND tc.table_name = $1
		) fk ON c.column_name = fk.column_name
		WHERE c.table_name = $1
		ORDER BY c.ordinal_position
	`

	rows, err := d.DB().QueryContext(ctx, columnQuery, tableName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var columns []models.Column
	for rows.Next() {
		var col models.Column
		var nullable string
		var defaultVal sql.NullString
		if err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal, &col.IsPrimaryKey, &col.IsForeignKey); err != nil {
			return nil, err
		}
		col.Nullable = nullable == "YES"
		if defaultVal.Valid {
			col.Default = defaultVal.String
		}
		columns = append(columns, col)
	}

	if len(columns) == 0 {
		return nil, ErrTableNotFound
	}

	ddlQuery := fmt.Sprintf("SELECT pg_get_ddl('public'::regnamespace, '%s'::regclass)", tableName)
	var ddl string
	if err := d.DB().QueryRowContext(ctx, ddlQuery).Scan(&ddl); err != nil {
		logging.Warn().
			Err(err).
			Str("table", tableName).
			Msg("Failed to retrieve DDL for table, continuing with empty DDL")
	}

	return &models.TableColumns{
		Columns: columns,
		DDL:     ddl,
	}, nil
}

func (d *PostgresDriver) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*models.QueryResult, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf("SELECT * FROM %q LIMIT %d OFFSET %d", tableName, limit, offset)
	return d.ExecuteQuery(ctx, query)
}

func (d *PostgresDriver) GetDatabases(ctx context.Context) ([]string, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	query := `
		SELECT datname
		FROM pg_database
		WHERE datistemplate = false
		ORDER BY datname
	`

	rows, err := d.DB().QueryContext(ctx, query)
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

func (d *PostgresDriver) GetVersion(ctx context.Context) (string, error) {
	if !d.IsConnected() || d.DB() == nil {
		return "", ErrNotConnected
	}

	var version string
	err := d.DB().QueryRowContext(ctx, "SELECT version()").Scan(&version)
	return version, err
}

func (d *PostgresDriver) GetServerInfo(ctx context.Context) (*models.ServerInfo, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	info := &models.ServerInfo{
		ServerType:     "PostgreSQL",
		AdditionalInfo: make(map[string]string),
	}

	// Get version
	if version, err := d.GetVersion(ctx); err == nil {
		info.Version = version
	}

	// Get current database and user
	_ = d.DB().QueryRowContext(ctx, "SELECT current_database()").Scan(&info.CurrentDatabase)
	_ = d.DB().QueryRowContext(ctx, "SELECT current_user").Scan(&info.CurrentUser)

	// Get uptime
	var uptime string
	if err := d.DB().QueryRowContext(ctx, "SELECT date_trunc('second', current_timestamp - pg_postmaster_start_time())::text").Scan(&uptime); err == nil {
		info.Uptime = uptime
	}

	// Get connection counts
	_ = d.DB().QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity").Scan(&info.ConnectionCount)

	// Get max connections
	_ = d.DB().QueryRowContext(ctx, "SELECT setting::int FROM pg_settings WHERE name = 'max_connections'").Scan(&info.MaxConnections)

	// Get database size
	var size string
	if err := d.DB().QueryRowContext(ctx, "SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&size); err == nil {
		info.DatabaseSize = size
	}

	// Additional PostgreSQL-specific info
	var serverEncoding string
	if err := d.DB().QueryRowContext(ctx, "SHOW server_encoding").Scan(&serverEncoding); err == nil {
		info.AdditionalInfo["Encoding"] = serverEncoding
	}

	var timezone string
	if err := d.DB().QueryRowContext(ctx, "SHOW timezone").Scan(&timezone); err == nil {
		info.AdditionalInfo["Timezone"] = timezone
	}

	return info, nil
}

func (d *PostgresDriver) GetQueryExecutionPlan(ctx context.Context, sql string) (*models.QueryResult, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	explainSQL := "EXPLAIN ANALYZE " + sql
	return d.ExecuteQuery(ctx, explainSQL)
}

func (d *PostgresDriver) GetTableIndexes(ctx context.Context, table string) ([]models.Index, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	query := d.buildIndexQuery()
	rows, err := d.DB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer func() { _ = rows.Close() }()

	return d.scanIndexRows(rows)
}

func (d *PostgresDriver) buildIndexQuery() string {

	return `
		SELECT 
			i.indexname,
			i.indexdef,
			ix.indisunique,
			ix.indisprimary,
			am.amname,
			ARRAY(
				SELECT a.attname
				FROM pg_attribute a
				WHERE a.attrelid = ix.indrelid
				AND a.attnum = ANY(ix.indkey)
				ORDER BY array_position(ix.indkey, a.attnum)
			) AS index_columns
		FROM pg_indexes i
		JOIN pg_class c ON c.relname = i.indexname
		JOIN pg_index ix ON ix.indexrelid = c.oid
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_am am ON am.oid = c.relam
		WHERE i.tablename = $1
		AND i.schemaname = current_schema()
		ORDER BY i.indexname
	`
}

func (d *PostgresDriver) scanIndexRows(rows *sql.Rows) ([]models.Index, error) {
	var indexes []models.Index
	for rows.Next() {
		index, err := d.scanIndexRow(rows)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, index)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return indexes, nil
}

func (d *PostgresDriver) scanIndexRow(rows *sql.Rows) (models.Index, error) {
	var indexName string
	var indexDef string
	var isUnique bool
	var isPrimary bool
	var accessMethod string
	var columns []string

	if err := rows.Scan(&indexName, &indexDef, &isUnique, &isPrimary, &accessMethod, pq.Array(&columns)); err != nil {
		return models.Index{}, fmt.Errorf("failed to scan index row: %w", err)
	}

	return models.Index{
		Name:      indexName,
		Columns:   columns,
		IsUnique:  isUnique,
		IsPrimary: isPrimary,
		Type:      accessMethod,
	}, nil
}

func (d *PostgresDriver) buildConnectionString(conn *models.Connection, secretsMgr secrets.Manager) (string, error) {
	var parts []string

	if conn.Host != "" {
		parts = append(parts, fmt.Sprintf("host=%s", conn.Host))
	}

	if conn.Port > 0 {
		parts = append(parts, fmt.Sprintf("port=%d", conn.Port))
	}

	if conn.Database != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", conn.Database))
	}

	if conn.Username != "" {
		parts = append(parts, fmt.Sprintf("user=%s", conn.Username))
	}

	if conn.SecretKeyID != "" && secretsMgr != nil {
		password, err := secretsMgr.RetrieveSecret(conn.SecretKeyID)
		if err != nil {
			logging.Warn().
				Err(err).
				Str("secretKeyID", conn.SecretKeyID).
				Msg("Failed to retrieve password from secrets manager, connection will proceed without stored password")
		} else if password != "" {
			parts = append(parts, fmt.Sprintf("password=%s", password))
		}
	}

	if conn.SSL != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", conn.SSL))
	} else {
		parts = append(parts, "sslmode=prefer")
	}

	if conn.SSLCACertPath != "" {

		certPath := strings.ReplaceAll(conn.SSLCACertPath, "\\", "/")
		parts = append(parts, fmt.Sprintf("sslrootcert=%s", certPath))
	}

	return strings.Join(parts, " "), nil
}
