package benthosbuilder_shared

// Holds the environment variable name and the connection id that should replace it at runtime when the Sync activity is launched
type BenthosDsn struct {
	EnvVarKey string
	// Neosync Connection Id
	ConnectionId string
}

type BenthosRedisConfig struct {
	Key    string
	Table  string // schema.table
	Column string
}
