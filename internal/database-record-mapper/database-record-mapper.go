package databaserecordmapper

import (
	"fmt"

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

func GetDatabaseRecordMapper(dbType string) (DatabaseRecordMapper[any], error) {
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
