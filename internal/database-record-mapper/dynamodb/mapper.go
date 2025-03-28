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

func (m *DynamoDBMapper) MapRecordWithKeyType(
	item map[string]types.AttributeValue,
) (valuemap map[string]any, typemap map[string]neosync_types.KeyType, err error) {
	standardJSON := make(map[string]any)
	ktm := make(map[string]neosync_types.KeyType)
	for k, v := range item {
		val, err := parseAttributeValue(k, v, ktm)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse attribute value for key %q: %w", k, err)
		}
		standardJSON[k] = val
	}
	return standardJSON, ktm, nil
}

// ParseAttributeValue converts a DynamoDB AttributeValue to a standard value
func parseAttributeValue(
	key string,
	v types.AttributeValue,
	keyTypeMap map[string]neosync_types.KeyType,
) (any, error) {
	switch t := v.(type) {
	case *types.AttributeValueMemberB:
		return t.Value, nil
	case *types.AttributeValueMemberBOOL:
		return t.Value, nil
	case *types.AttributeValueMemberBS:
		return t.Value, nil
	case *types.AttributeValueMemberL:
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			val, err := parseAttributeValue(fmt.Sprintf("%s[%d]", key, i), v, keyTypeMap)
			if err != nil {
				return nil, fmt.Errorf("failed to parse list value at index %d for key %q: %w", i, key, err)
			}
			lAny[i] = val
		}
		return lAny, nil
	case *types.AttributeValueMemberM:
		mAny := make(map[string]any, len(t.Value))
		for k, v := range t.Value {
			path := k
			if key != "" {
				path = fmt.Sprintf("%s.%s", key, k)
			}
			val, err := parseAttributeValue(path, v, keyTypeMap)
			if err != nil {
				return nil, fmt.Errorf("failed to parse map value for key %q: %w", path, err)
			}
			mAny[k] = val
		}
		return mAny, nil
	case *types.AttributeValueMemberN:
		n, err := gotypeutil.ParseStringAsNumber(t.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse number value for key %q: %w", key, err)
		}
		return n, nil
	case *types.AttributeValueMemberNS:
		keyTypeMap[key] = neosync_types.NumberSet
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			n, err := gotypeutil.ParseStringAsNumber(v)
			if err != nil {
				return nil, fmt.Errorf("failed to parse number set value at index %d for key %q: %w", i, key, err)
			}
			lAny[i] = n
		}
		return lAny, nil
	case *types.AttributeValueMemberNULL:
		return nil, nil
	case *types.AttributeValueMemberS:
		return t.Value, nil
	case *types.AttributeValueMemberSS:
		keyTypeMap[key] = neosync_types.StringSet
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			lAny[i] = v
		}
		return lAny, nil
	}
	return nil, fmt.Errorf("unsupported DynamoDB attribute type for key %q", key)
}
