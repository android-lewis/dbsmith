package constants

import "time"

const (
	DefaultMaxResults = 10000
	DefaultTimeout    = 30 * time.Second
	PingTimeout       = 5 * time.Second
)

const (
	DefaultLogFileName       = "dbsmith.log"
	DefaultConfigFileName    = "config.log"
	DefaultWorkspaceFileName = "workspace.yaml"
	DefaultMaxSizeMB         = 10
	DefaultMaxBackups        = 3
	DefaultMaxAgeDays        = 28
)

const (
	WorkspaceVersion = 1
)

var SupportedDrivers = []string{"postgres", "mysql", "sqlite"}
