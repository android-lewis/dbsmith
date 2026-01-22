# DBSmith

A terminal-based database client for PostgreSQL, MySQL, and SQLite.

This is in very early development, I plan to support more DBs, autocompletion, server statistics and much more...

## Features

- **Multi-Database Support**: PostgreSQL, MySQL/MariaDB, SQLite
- **Schema Explorer**: Browse schemas, tables, columns, and indexes
- **SQL Editor**: Multi-tab editor with syntax highlighting
- **Query Management**: Save, load, and organise queries
- **Export**: CSV, JSON
- **Secure Credentials**: System keyring integration with encrypted file fallback
- **Workspace Persistence**: YAML-based workspace files for connections and queries

## Installation

### Build from Source

```bash
git clone https://github.com/android-lewis/dbsmith.git
cd dbsmith
go build -o dbsmith ./cmd/dbsmith
```

### Run

```bash
./dbsmith
```

## Usage

DBSmith launches directly into a TUI. On first run, you'll be prompted to create or load a workspace file.

## Configuration

Workspaces are stored as YAML files (default: `~/.config/dbsmith/workspace.yaml`):
Passwords are stored in the system keyring when available, otherwise in an encrypted file at `~/.config/dbsmith/.secrets`.

