package databaserecordmapper

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type DynamoDBMapper struct{}

func NewDynamoBuilder() *Builder[map[string]types.AttributeValue] {
	return &Builder[map[string]types.AttributeValue]{
		mapper: &DynamoDBMapper{},
	}
}

func (m *DynamoDBMapper) MapRecord(item map[string]types.AttributeValue) (map[string]any, error) {
	return nil, errors.ErrUnsupported
}

func (m *DynamoDBMapper) MapRecordWithKeyType(item map[string]types.AttributeValue) (map[string]any, map[string]neosync_types.KeyType, error) {
	return nil, nil, nil
}
