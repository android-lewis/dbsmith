package constants

import "errors"

var (
	ErrQueryTimeout   = errors.New("query execution timeout")
	ErrQueryCancelled = errors.New("query cancelled by user")
	ErrNotConnected   = errors.New("database not connected")
)
