package databaserecordmapper

import (
	"errors"

	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type MongoDBMapper struct{}

func NewMongoBuilder() *Builder[map[string]any] {
	return &Builder[map[string]any]{
		mapper: &MongoDBMapper{},
	}
}

func (m *MongoDBMapper) MapRecord(item map[string]any) (map[string]any, error) {
	return nil, errors.ErrUnsupported
}

func (m *MongoDBMapper) MapRecordWithKeyType(item map[string]any) (valuemap map[string]any, typemap map[string]neosync_types.KeyType, err error) {
	return nil, nil, nil
}
