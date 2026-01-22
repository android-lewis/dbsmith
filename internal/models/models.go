package models

import "time"

type ConnectionType string

const (
	PostgresType ConnectionType = "postgres"
	MySQLType    ConnectionType = "mysql"
	SQLiteType   ConnectionType = "sqlite"
)

type Connection struct {
	Name          string         `yaml:"name"`
	Type          ConnectionType `yaml:"type"`
	Host          string         `yaml:"host,omitempty"`
	Port          int            `yaml:"port,omitempty"`
	Database      string         `yaml:"database,omitempty"`
	Username      string         `yaml:"username,omitempty"`
	SecretKeyID   string         `yaml:"secret_key_id,omitempty"`
	SSL           string         `yaml:"ssl,omitempty"`
	SSLCACertPath string         `yaml:"ssl_ca_cert_path,omitempty"`
	CreatedAt     time.Time      `yaml:"created_at,omitempty"`
	LastModified  time.Time      `yaml:"last_modified,omitempty"`
}

func (c *Connection) GetSQLDialect() string {
	switch c.Type {
	case PostgresType:
		return "postgresql"
	case MySQLType:
		return "mysql"
	case SQLiteType:
		return "sqlite"
	default:
		return "sql"
	}
}

type Workspace struct {
	Name         string           `yaml:"name"`
	Connections  []Connection     `yaml:"connections"`
	SavedQueries []SavedQuery     `yaml:"saved_queries"`
	LastOpenTabs []string         `yaml:"last_open_tabs,omitempty"`
	Preferences  *UserPreferences `yaml:"preferences,omitempty"`
	CreatedAt    time.Time        `yaml:"created_at,omitempty"`
	LastModified time.Time        `yaml:"last_modified,omitempty"`
	Version      int              `yaml:"version,omitempty"`
}

type UserPreferences struct {
	AutocompleteEnabled bool `yaml:"autocomplete_enabled"`
}

type SavedQuery struct {
	ID           string    `yaml:"id"`
	Name         string    `yaml:"name"`
	SQL          string    `yaml:"sql"`
	Description  string    `yaml:"description,omitempty"`
	CreatedAt    time.Time `yaml:"created_at,omitempty"`
	LastModified time.Time `yaml:"last_modified,omitempty"`
}

type ExecutionRecord struct {
	QueryID      string
	Query        string
	ExecutedAt   time.Time
	Duration     int64
	RowsAffected int64
	Error        string
}

type Schema struct {
	Name  string
	Owner string
}

type Table struct {
	Name   string
	Schema string
	Type   string
}

type Column struct {
	Name         string
	Type         string
	Nullable     bool
	Default      string
	IsPrimaryKey bool
	IsForeignKey bool
	Description  string
}

type Index struct {
	Name      string
	Columns   []string
	IsUnique  bool
	IsPrimary bool
	Type      string
}

type TableColumns struct {
	Columns []Column
	DDL     string
}

type QueryResult struct {
	Columns     []string
	ColumnTypes []string
	Rows        [][]interface{}
	RowCount    int64
	ExecutionMs int64
	Error       error
}

type ExportConfig struct {
	Format        string
	Destination   string
	IncludeHeader bool
}
