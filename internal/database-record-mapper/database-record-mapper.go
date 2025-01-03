package databaserecordmapper

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type DatabaseRecordMapper[T any] interface {
	// MapRecord returns a map where:
	// - Keys are column/field names from the database record
	// - Values are either Go native types (string, int, etc.) or Neosync custom types
	MapRecord(record T) (map[string]any, error)

	// Deprecated: use MapRecord instead with neosync types
	MapRecordWithKeyType(record T) (map[string]any, map[string]neosync_types.KeyType, error)
}

type Builder[T any] struct {
	mapper DatabaseRecordMapper[T]
}

func (b *Builder[T]) MapRecord(record any) (map[string]any, error) {
	typedRecord, ok := record.(T)
	if !ok {
		return nil, fmt.Errorf("invalid record type: expected %T, got %T", *new(T), record)
	}
	return b.mapper.MapRecord(typedRecord)
}

// Deprecated: use MapRecord instead with neosync types
func (b *Builder[T]) MapRecordWithKeyType(record any) (map[string]any, map[string]neosync_types.KeyType, error) {
	typedRecord, ok := record.(T)
	if !ok {
		return nil, nil, fmt.Errorf("invalid record type: expected %T, got %T", *new(T), record)
	}
	return b.mapper.MapRecordWithKeyType(typedRecord)
}

func NewDatabaseRecordMapper(dbType string) (DatabaseRecordMapper[any], error) {
	switch dbType {
	case sqlmanager_shared.PostgresDriver:
		return NewPostgresBuilder(), nil
	case sqlmanager_shared.MysqlDriver:
		return NewMySQLBuilder(), nil
	case sqlmanager_shared.MssqlDriver:
		return NewMSSQLBuilder(), nil
	case "dynamodb":
		return NewDynamoBuilder(), nil
	case "mongodb":
		return NewMongoBuilder(), nil
	default:
		return nil, fmt.Errorf("database type %s not supported", dbType)
	}
}

func NewDatabaseRecordMapperFromConnection(connection *mgmtv1alpha1.Connection) (DatabaseRecordMapper[any], error) {
	switch connection.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return NewDatabaseRecordMapper(sqlmanager_shared.PostgresDriver)
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return NewDatabaseRecordMapper(sqlmanager_shared.MysqlDriver)
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return NewDatabaseRecordMapper(sqlmanager_shared.MssqlDriver)
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return NewDatabaseRecordMapper("mongodb")
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		return NewDatabaseRecordMapper("dynamodb")
	default:
		return nil, fmt.Errorf("unsupported connection type: %T for database record mapper", connection.GetConnectionConfig().GetConfig())
	}
}
