package db

import "errors"

var (
	ErrNotConnected         = errors.New("not connected to database")
	ErrConnectionFailed     = errors.New("failed to connect to database")
	ErrInvalidConnection    = errors.New("invalid connection parameters")
	ErrUnsupportedDriver    = errors.New("unsupported database driver")
	ErrQueryFailed          = errors.New("query execution failed")
	ErrTableNotFound        = errors.New("table not found")
	ErrDatabaseNotFound     = errors.New("database not found")
	ErrInvalidSQL           = errors.New("invalid SQL syntax")
	ErrOperationTimeout     = errors.New("operation timeout")
	ErrUnsupportedOperation = errors.New("this operation is not supported")
	ErrInvalidIdentifier    = errors.New("invalid identifier")
)
