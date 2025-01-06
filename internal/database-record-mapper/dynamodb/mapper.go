package dynamodb

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nucleuscloud/neosync/internal/database-record-mapper/builder"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type DynamoDBMapper struct{}

func NewDynamoBuilder() *builder.Builder[map[string]types.AttributeValue] {
	return &builder.Builder[map[string]types.AttributeValue]{
		Mapper: &DynamoDBMapper{},
	}
}

func (m *DynamoDBMapper) MapRecord(item map[string]types.AttributeValue) (map[string]any, error) {
	return nil, errors.ErrUnsupported
}

func (m *DynamoDBMapper) MapRecordWithKeyType(item map[string]types.AttributeValue) (valuemap map[string]any, typemap map[string]neosync_types.KeyType, err error) {
	standardJSON := make(map[string]any)
	ktm := make(map[string]neosync_types.KeyType)
	for k, v := range item {
		val := parseAttributeValue(k, v, ktm)
		standardJSON[k] = val
	}
	return standardJSON, ktm, nil
}

// ParseAttributeValue converts a DynamoDB AttributeValue to a standard value
func parseAttributeValue(key string, v types.AttributeValue, keyTypeMap map[string]neosync_types.KeyType) any {
	switch t := v.(type) {
	case *types.AttributeValueMemberB:
		return t.Value
	case *types.AttributeValueMemberBOOL:
		return t.Value
	case *types.AttributeValueMemberBS:
		return t.Value
	case *types.AttributeValueMemberL:
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			val := parseAttributeValue(fmt.Sprintf("%s[%d]", key, i), v, keyTypeMap)
			lAny[i] = val
		}
		return lAny
	case *types.AttributeValueMemberM:
		mAny := make(map[string]any, len(t.Value))
		for k, v := range t.Value {
			path := k
			if key != "" {
				path = fmt.Sprintf("%s.%s", key, k)
			}
			val := parseAttributeValue(path, v, keyTypeMap)
			mAny[k] = val
		}
		return mAny
	case *types.AttributeValueMemberN:
		n, err := gotypeutil.ParseStringAsNumber(t.Value)
		if err != nil {
			return t.Value
		}
		return n
	case *types.AttributeValueMemberNS:
		keyTypeMap[key] = neosync_types.NumberSet
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			n, err := gotypeutil.ParseStringAsNumber(v)
			if err != nil {
				return v
			}
			lAny[i] = n
		}
		return lAny
	case *types.AttributeValueMemberNULL:
		return nil
	case *types.AttributeValueMemberS:
		return t.Value
	case *types.AttributeValueMemberSS:
		keyTypeMap[key] = neosync_types.StringSet
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			lAny[i] = v
		}
		return lAny
	}
	return nil
}
