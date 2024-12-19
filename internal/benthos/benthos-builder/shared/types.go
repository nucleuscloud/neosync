package benthosbuilder_shared

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
)

// Holds the environment variable name and the connection id that should replace it at runtime when the Sync activity is launched
type BenthosDsn struct {
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
		runConfigs []*tabledependency.RunConfig,
		subsetByForeignKeyConstraints bool,
		groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	) (map[string]map[tabledependency.RunType]*sqlmanager_shared.SelectQuery, error)
}

func WithEnvInterpolation(input string) string {
	return fmt.Sprintf("${%s}", input)
}

// ConnectionType represents supported connection types
type ConnectionType string

const (
	ConnectionTypePostgres    ConnectionType = "postgres"
	ConnectionTypeMysql       ConnectionType = "mysql"
	ConnectionTypeMssql       ConnectionType = "mssql"
	ConnectionTypeAwsS3       ConnectionType = "aws-s3"
	ConnectionTypeGCP         ConnectionType = "gcp-cloud-storage"
	ConnectionTypeMongo       ConnectionType = "mongodb"
	ConnectionTypeDynamodb    ConnectionType = "aws-dynamodb"
	ConnectionTypeLocalDir    ConnectionType = "local-directory"
	ConnectionTypeOpenAI      ConnectionType = "openai"
	ConnectionTypeNeosyncData ConnectionType = "neosync-data-stream"
)

// Determines type of connection from Connection
func GetConnectionType(connection *mgmtv1alpha1.Connection) (ConnectionType, error) {
	switch connection.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return ConnectionTypePostgres, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return ConnectionTypeMysql, nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return ConnectionTypeMssql, nil
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		return ConnectionTypeAwsS3, nil
	case *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig:
		return ConnectionTypeGCP, nil
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return ConnectionTypeMongo, nil
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		return ConnectionTypeDynamodb, nil
	case *mgmtv1alpha1.ConnectionConfig_LocalDirConfig:
		return ConnectionTypeLocalDir, nil
	case *mgmtv1alpha1.ConnectionConfig_OpenaiConfig:
		return ConnectionTypeOpenAI, nil
	default:
		return "", fmt.Errorf("unsupported connection type: %T", connection.GetConnectionConfig().GetConfig())
	}
}
