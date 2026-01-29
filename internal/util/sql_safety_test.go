package util

import "testing"

func TestAnalyzeQuerySafety(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		wantDestructive bool
		wantType      string
	}{
		// Safe queries
		{"SELECT", "SELECT * FROM users", false, ""},
		{"SELECT with WHERE", "SELECT * FROM users WHERE id = 1", false, ""},
		{"INSERT", "INSERT INTO users (name) VALUES ('test')", false, ""},
		{"DELETE with WHERE", "DELETE FROM users WHERE id = 1", false, ""},
		{"UPDATE with WHERE", "UPDATE users SET name = 'test' WHERE id = 1", false, ""},
		{"CREATE TABLE", "CREATE TABLE test (id INT)", false, ""},

		// Destructive queries
		{"DROP TABLE", "DROP TABLE users", true, "DROP"},
		{"DROP DATABASE", "DROP DATABASE mydb", true, "DROP"},
		{"DROP INDEX", "DROP INDEX idx_users", true, "DROP"},
		{"TRUNCATE", "TRUNCATE TABLE users", true, "TRUNCATE"},
		{"TRUNCATE no TABLE keyword", "TRUNCATE users", true, "TRUNCATE"},
		{"DELETE without WHERE", "DELETE FROM users", true, "DELETE"},
		{"UPDATE without WHERE", "UPDATE users SET active = false", true, "UPDATE"},

		// Case insensitivity
		{"drop lowercase", "drop table users", true, "DROP"},
		{"delete lowercase", "delete from users", true, "DELETE"},
		{"update lowercase", "update users set name = 'x'", true, "UPDATE"},
		{"truncate lowercase", "truncate users", true, "TRUNCATE"},

		// With comments
		{"DELETE with line comment", "-- comment\nDELETE FROM users", true, "DELETE"},
		{"DELETE with block comment", "/* comment */ DELETE FROM users", true, "DELETE"},
		{"DELETE with WHERE after comment", "DELETE FROM users /* comment */ WHERE id = 1", false, ""},

		// Edge cases
		{"DELETE with WHERE in comment", "DELETE FROM users -- WHERE id = 1", true, "DELETE"},
		{"UPDATE with subquery WHERE", "UPDATE users SET x = 1 WHERE id IN (SELECT id FROM t)", false, ""},
		{"Multiline DELETE with WHERE", "DELETE\nFROM\nusers\nWHERE\nid = 1", false, ""},
		{"Empty string", "", false, ""},
		{"Whitespace only", "   \n\t  ", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeQuerySafety(tt.sql)

			if result.IsDestructive != tt.wantDestructive {
				t.Errorf("IsDestructive = %v, want %v", result.IsDestructive, tt.wantDestructive)
			}

			if tt.wantDestructive && result.QueryType != tt.wantType {
				t.Errorf("QueryType = %q, want %q", result.QueryType, tt.wantType)
			}

			if tt.wantDestructive && result.Warning == "" {
				t.Error("expected non-empty warning for destructive query")
			}
		})
	}
}

func TestCleanSQL(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want string
	}{
		{"simple", "SELECT 1", "SELECT 1"},
		{"line comment", "SELECT 1 -- comment", "SELECT 1"},
		{"block comment", "SELECT /* comment */ 1", "SELECT   1"}, // block comments become single space
		{"multiline", "SELECT\n1", "SELECT 1"},
		{"multiple line comments", "-- first\nSELECT 1\n-- second", "SELECT 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanSQL(tt.sql)
			if got != tt.want {
				t.Errorf("cleanSQL() = %q, want %q", got, tt.want)
			}
		})
	}
}
