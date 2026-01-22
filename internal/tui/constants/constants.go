package constants

import "time"

const (
	TimeoutAutocomplete = 2 * time.Second
	TimeoutSchemaLoad   = 3 * time.Second
	TimeoutQueryExec    = 30 * time.Second
	TimeoutConnection   = 5 * time.Second
)

const (
	MaxCellDisplayLen      = 100
	MaxPreviewCellLen      = 50
	DefaultDataLimit       = 100
	ProgressUpdateInterval = 100
	LargeResultThreshold   = 1000
	DefaultBatchSize       = 100
	DefaultCellCacheSize   = 4000
)
