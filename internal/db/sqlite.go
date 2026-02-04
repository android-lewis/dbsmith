package db

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/android-lewis/dbsmith/internal/logging"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/secrets"
	_ "modernc.org/sqlite"
)

type SQLiteDriver struct {
	BaseDriver
}

func NewSQLiteDriver() *SQLiteDriver {
	return &SQLiteDriver{}
}

func validateSQLiteIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_$]*$`)
	if !pattern.MatchString(identifier) {
		return fmt.Errorf("invalid identifier: contains unsafe characters")
	}
	return nil
}

func (d *SQLiteDriver) Connect(ctx context.Context, conn *models.Connection, secretsMgr secrets.Manager) error {
	if err := d.validateConnection(conn, models.SQLiteType); err != nil {
		return err
	}

	if conn.Database == "" {
		return ErrInvalidConnection
	}

	return d.ConnectWithDSN(ctx, "sqlite", conn.Database, conn)
}

func (d *SQLiteDriver) GetSchemas(ctx context.Context) ([]models.Schema, error) {
	return []models.Schema{
		{
			Name:  "",
			Owner: "",
		},
	}, nil
}

func (d *SQLiteDriver) GetTables(ctx context.Context, schema models.Schema) ([]models.Table, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"

	rows, err := d.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer func() { _ = rows.Close() }()

	var tables []models.Table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		table := models.Table{
			Name:   tableName,
			Schema: schema.Name,
			Type:   "BASE TABLE",
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

func (d *SQLiteDriver) GetTableColumns(ctx context.Context, tableName string) (*models.TableColumns, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	tableName = strings.TrimSpace(tableName)

	if err := validateSQLiteIdentifier(tableName); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	fkColumns := make(map[string]bool)
	fkQuery := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	fkRows, err := d.DB().QueryContext(ctx, fkQuery)
	if err == nil {
		defer func() { _ = fkRows.Close() }()
		for fkRows.Next() {
			var id, seq int
			var table, from, to, onUpdate, onDelete, match string
			if err := fkRows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); err == nil {
				fkColumns[from] = true
			}
		}
	}

	columnQuery := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := d.DB().QueryContext(ctx, columnQuery)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var columns []models.Column
	for rows.Next() {
		var col models.Column
		var cid int
		var notnull int
		var dfltValue interface{}
		var pk int

		if err := rows.Scan(&cid, &col.Name, &col.Type, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}

		col.Nullable = notnull == 0
		col.IsPrimaryKey = pk == 1
		col.IsForeignKey = fkColumns[col.Name]

		if dfltValue != nil {
			col.Default = fmt.Sprintf("%v", dfltValue)
		}

		columns = append(columns, col)
	}

	if len(columns) == 0 {
		return nil, ErrTableNotFound
	}

	var ddl string
	ddlQuery := "SELECT sql FROM sqlite_master WHERE type='table' AND name=?"
	if err := d.DB().QueryRowContext(ctx, ddlQuery, tableName).Scan(&ddl); err != nil {
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

func (d *SQLiteDriver) GetTableIndexes(ctx context.Context, table string) ([]models.Index, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	table = strings.TrimSpace(table)

	if err := validateSQLiteIdentifier(table); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	indexListQuery := fmt.Sprintf("PRAGMA index_list(%s)", table)
	rows, err := d.DB().QueryContext(ctx, indexListQuery)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer func() { _ = rows.Close() }()

	var indexes []models.Index
	for rows.Next() {
		var seq int
		var name string
		var unique int
		var origin string
		var partial int

		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, err
		}

		index := models.Index{
			Name:      name,
			IsUnique:  unique == 1,
			IsPrimary: origin == "pk",
			Type:      "BTREE",
		}

		if err := validateSQLiteIdentifier(name); err != nil {
			return nil, fmt.Errorf("invalid index name: %w", err)
		}
		indexInfoQuery := fmt.Sprintf("PRAGMA index_info(%s)", name)
		colRows, err := d.DB().QueryContext(ctx, indexInfoQuery)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}

		var columns []string
		for colRows.Next() {
			var seqno int
			var cid int
			var colName string

			if err := colRows.Scan(&seqno, &cid, &colName); err != nil {
				_ = colRows.Close()
				return nil, err
			}
			columns = append(columns, colName)
		}
		if closeErr := colRows.Close(); closeErr != nil {
			return nil, closeErr
		}

		if err := colRows.Err(); err != nil {
			return nil, err
		}

		index.Columns = columns
		indexes = append(indexes, index)
	}

	return indexes, rows.Err()
}

func (d *SQLiteDriver) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*models.QueryResult, error) {
	if !d.IsConnected() || d.DB() == nil {
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

func (d *SQLiteDriver) GetDatabases(ctx context.Context) ([]string, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	return []string{"main"}, nil
}

func (d *SQLiteDriver) GetVersion(ctx context.Context) (string, error) {
	if !d.IsConnected() || d.DB() == nil {
		return "", ErrNotConnected
	}

	var version string
	err := d.DB().QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version)
	return version, err
}

func (d *SQLiteDriver) GetServerInfo(ctx context.Context) (*models.ServerInfo, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	info := &models.ServerInfo{
		ServerType:     "SQLite",
		AdditionalInfo: make(map[string]string),
	}

	// Get version
	if version, err := d.GetVersion(ctx); err == nil {
		info.Version = version
	}

	// SQLite uses file path as "database"
	if d.connection != nil && d.connection.Database != "" {
		info.CurrentDatabase = d.connection.Database
	}

	// SQLite is embedded, no connection/user concepts
	info.CurrentUser = "N/A (embedded)"
	info.ConnectionCount = 1
	info.MaxConnections = 1

	// Get database size (page_count * page_size)
	var pageCount, pageSize int64
	if err := d.DB().QueryRowContext(ctx, "PRAGMA page_count").Scan(&pageCount); err == nil {
		if err := d.DB().QueryRowContext(ctx, "PRAGMA page_size").Scan(&pageSize); err == nil {
			sizeBytes := pageCount * pageSize
			if sizeBytes < 1024 {
				info.DatabaseSize = fmt.Sprintf("%d B", sizeBytes)
			} else if sizeBytes < 1024*1024 {
				info.DatabaseSize = fmt.Sprintf("%.2f KB", float64(sizeBytes)/1024)
			} else {
				info.DatabaseSize = fmt.Sprintf("%.2f MB", float64(sizeBytes)/(1024*1024))
			}
		}
	}

	// SQLite-specific info
	var journalMode string
	if err := d.DB().QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err == nil {
		info.AdditionalInfo["Journal Mode"] = journalMode
	}

	var encoding string
	if err := d.DB().QueryRowContext(ctx, "PRAGMA encoding").Scan(&encoding); err == nil {
		info.AdditionalInfo["Encoding"] = encoding
	}

	var autoVacuum int
	if err := d.DB().QueryRowContext(ctx, "PRAGMA auto_vacuum").Scan(&autoVacuum); err == nil {
		vacuumMode := "None"
		switch autoVacuum {
		case 1:
			vacuumMode = "Full"
		case 2:
			vacuumMode = "Incremental"
		}
		info.AdditionalInfo["Auto Vacuum"] = vacuumMode
	}

	return info, nil
}

func (d *SQLiteDriver) GetQueryExecutionPlan(ctx context.Context, sql string) (*models.QueryResult, error) {
	if !d.IsConnected() || d.DB() == nil {
		return nil, ErrNotConnected
	}

	explainSQL := "EXPLAIN QUERY PLAN " + sql
	return d.ExecuteQuery(ctx, explainSQL)
}
