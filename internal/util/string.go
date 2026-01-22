package util

import "github.com/android-lewis/dbsmith/internal/constants"

func DialectDisplayName(dialect string) string {
	switch dialect {
	case "postgres", "postgresql":
		return "PostgreSQL"
	case "mysql":
		return "MySQL"
	case "sqlite", "sqlite3":
		return "SQLite"
	default:
		return "SQL"
	}
}

func IsSupportedDriver(driver string) bool {
	for _, d := range constants.SupportedDrivers {
		if d == driver {
			return true
		}
	}
	return false
}
