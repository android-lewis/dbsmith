# dbsmith - Golang TUI Database Workstation

A modern Terminal User Interface (TUI) application for exploring, querying, and exporting data from multiple relational databases. Written in Go with support for PostgreSQL, MySQL/MariaDB, and SQLite.

## Features (Planned v0.1)

- ✅ **Multi-Database Support**: PostgreSQL, MySQL/MariaDB, SQLite
- ✅ **Schema Explorer**: Browse databases, schemas, tables, columns, and indexes
- ✅ **SQL Editor**: Write and execute SQL queries with syntax highlighting
- ✅ **Streaming Results**: Handle large result sets efficiently (10M+ rows)
- ✅ **Export Formats**: CSV, TSV, JSON, JSONL, SQL INSERT statements
- ✅ **Workspace Management**: Save connections and queries for reuse
- ✅ **Secure Credentials**: Passwords stored in OS keyring or encrypted vault
- ✅ **CLI Mode**: Headless query and export commands
- ✅ **Cross-Platform**: macOS, Linux, Windows

## Project Structure

```
cmd/
├── dbsmith/              # Application entry point
internal/
├── cmd/                  # CLI command implementations (Cobra)
├── tui/                  # Terminal UI components (Bubble Tea)
├── db/                   # Database drivers (PostgreSQL, MySQL, SQLite)
├── workspace/            # Workspace persistence (YAML)
├── secrets/              # Credential storage (keyring/encrypted)
├── exporter/             # Export format implementations
└── executor/             # Query execution engine
pkg/
└── models/               # Shared data structures
```

## Quick Start (Development)

### Prerequisites
- Go 1.21 or later
- Git
- [Task](https://taskfile.dev/installation/) (task runner)

### Build
```bash
git clone https://github.com/user/dbsmith.git
cd dbsmith
task build
./bin/dbsmith --help
```

### Run Tests
```bash
task test
```

### Build for All Platforms
```bash
task build:all
```

### See All Available Tasks
```bash
task --list-all
```

## Commands

### TUI Mode
```bash
dbsmith open [workspace-file]
```
Opens the interactive terminal UI.

### Query Mode
```bash
dbsmith query <connection> <sql>
dbsmith query my_pg_conn "SELECT * FROM users LIMIT 10"
```

### Export Mode
```bash
dbsmith export <connection> <sql> <output-file>
dbsmith export my_pg_conn "SELECT * FROM users" users.csv --format csv
```

### Workspace Management
```bash
dbsmith list-connections
dbsmith store-secret my_pg_password
```

### Debugging
```bash
dbsmith debug          # System info and diagnostics
dbsmith version        # Version info
```

## Dependencies

- **Charmbracelet Bubbletea**: TUI framework
- **Charmbracelet Lipgloss**: Terminal styling
- **Cobra**: CLI framework
- **go-keyring**: OS keyring integration
- **lib/pq**: PostgreSQL driver
- **go-sql-driver/mysql**: MySQL driver
- **mattn/go-sqlite3**: SQLite driver
- **gopkg.in/yaml.v3**: Workspace serialization

## Contributing

(To be defined)

## License

(To be defined)

## Support

For issues, questions, or suggestions, please open an issue on GitHub.
