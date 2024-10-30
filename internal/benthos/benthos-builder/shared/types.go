package benthosbuilder_shared

import (
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
)

// Holds the environment variable name and the connection id that should replace it at runtime when the Sync activity is launched
type BenthosDsn struct {
	EnvVarKey string
	// Neosync Connection Id
	ConnectionId string
}

// Keeps track of redis keys for clean up after syncing a table
type BenthosRedisConfig struct {
	Key    string
	Table  string // schema.table
	Column string
}

// querybuilder wrapper to avoid cgo in the cli
type SelectQueryMapBuilder interface {
	BuildSelectQueryMap(
		driver string,
		tableFkConstraints map[string][]*sqlmanager_shared.ForeignConstraint,
		runConfigs []*tabledependency.RunConfig,
		subsetByForeignKeyConstraints bool,
		groupedColumnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
	) (map[string]map[tabledependency.RunType]string, error)
}
