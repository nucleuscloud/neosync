package dynamodb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	gotypeutil "github.com/nucleuscloud/neosync/internal/gotypeutil"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

func UnmarshalDynamoDBItem(item map[string]any) (standardMap map[string]any, keyTypeMap map[string]neosync_types.KeyType) {
	result := make(map[string]any)
	ktm := make(map[string]neosync_types.KeyType)
	for key, value := range item {
		result[key] = ParseDynamoDBAttributeValue(key, value, ktm)
	}

	return result, ktm
}

func ParseDynamoDBAttributeValue(key string, value any, keyTypeMap map[string]neosync_types.KeyType) any {
	if m, ok := value.(map[string]any); ok {
		for dynamoType, dynamoValue := range m {
			switch dynamoType {
			case "S":
				return dynamoValue.(string)
			case "B":
				switch v := dynamoValue.(type) {
				case string:
					byteSlice, err := base64.StdEncoding.DecodeString(v)
					if err != nil {
						return dynamoValue
					}
					return byteSlice
				case []byte:
					return v
				default:
					return dynamoValue
				}
			case "N":
				n, err := gotypeutil.ParseStringAsNumber(dynamoValue.(string))
				if err != nil {
					return dynamoValue
				}
				return n
			case "BOOL":
				return dynamoValue.(bool)
			case "NULL":
				return nil
			case "L":
				list := dynamoValue.([]any)
				result := make([]any, len(list))
				for i, item := range list {
					result[i] = ParseDynamoDBAttributeValue(fmt.Sprintf("%s[%d]", key, i), item, keyTypeMap)
				}
				return result
			case "M":
				mAny := map[string]any{}
				for k, v := range dynamoValue.(map[string]any) {
					path := k
					if key != "" {
						path = fmt.Sprintf("%s.%s", key, k)
					}
					val := ParseDynamoDBAttributeValue(path, v, keyTypeMap)
					mAny[k] = val
				}
				return mAny
			case "BS":
				var result [][]byte
				switch bytes := dynamoValue.(type) {
				case []any:
					result = make([][]byte, len(bytes))
					for i, b := range bytes {
						s := b.(string)
						byteSlice, err := base64.StdEncoding.DecodeString(s)
						if err != nil {
							return dynamoValue
						}
						result[i] = byteSlice
					}
				case [][]byte:
					return bytes
				default:
					return dynamoValue
				}
				return result
			case "SS":
				keyTypeMap[key] = neosync_types.StringSet
				switch ss := dynamoValue.(type) {
				case []any:
					result := make([]string, len(ss))
					for i, s := range ss {
						result[i] = s.(string)
					}
					return result
				case []string:
					return ss
				default:
					return dynamoValue
				}
			case "NS":
				keyTypeMap[key] = neosync_types.NumberSet
				var result []any
				switch ns := dynamoValue.(type) {
				case []any:
					result = make([]any, len(ns))
					for i, num := range ns {
						n, err := gotypeutil.ParseStringAsNumber(num.(string))
						if err != nil {
							result[i] = num
						}
						result[i] = n
					}
				case []string:
					result = make([]any, len(ns))
					for i, num := range ns {
						n, err := gotypeutil.ParseStringAsNumber(num)
						if err != nil {
							result[i] = num
						}
						result[i] = n
					}
				default:
					return dynamoValue
				}
				return result
			}
		}
	}
	return value
}

func UnmarshalAttributeValueMap(item map[string]types.AttributeValue) (standardMap map[string]any, keyTypeMap map[string]neosync_types.KeyType) {
	standardJSON := make(map[string]any)
	ktm := make(map[string]neosync_types.KeyType)
	for k, v := range item {
		val := ParseAttributeValue(k, v, ktm)
		standardJSON[k] = val
	}
	return standardJSON, ktm
}

// ParseAttributeValue converts a DynamoDB AttributeValue to a standard value
func ParseAttributeValue(key string, v types.AttributeValue, keyTypeMap map[string]neosync_types.KeyType) any {
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
			val := ParseAttributeValue(fmt.Sprintf("%s[%d]", key, i), v, keyTypeMap)
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
			val := ParseAttributeValue(path, v, keyTypeMap)
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

func MarshalToAttributeValue(key string, root any, keyTypeMap map[string]neosync_types.KeyType) types.AttributeValue {
	if typeStr, ok := keyTypeMap[key]; ok {
		switch typeStr {
		case neosync_types.StringSet:
			s, err := ConvertToStringSlice(root)
			if err == nil {
				return &types.AttributeValueMemberSS{
					Value: s,
				}
			}
		case neosync_types.NumberSet:
			s, err := ConvertToStringSlice(root)
			if err == nil {
				return &types.AttributeValueMemberNS{
					Value: s,
				}
			}
		}
	}
	switch v := root.(type) {
	case map[string]any:
		m := make(map[string]types.AttributeValue, len(v))
		for k, v2 := range v {
			path := k
			if key != "" {
				path = fmt.Sprintf("%s.%s", key, k)
			}
			m[k] = MarshalToAttributeValue(path, v2, keyTypeMap)
		}
		return &types.AttributeValueMemberM{
			Value: m,
		}
	case []byte:
		return &types.AttributeValueMemberB{
			Value: v,
		}
	case [][]byte:
		return &types.AttributeValueMemberBS{
			Value: v,
		}
	case []any:
		l := make([]types.AttributeValue, len(v))
		for i, v2 := range v {
			l[i] = MarshalToAttributeValue(fmt.Sprintf("%s[%d]", key, i), v2, keyTypeMap)
		}
		return &types.AttributeValueMemberL{
			Value: l,
		}
	case string:
		return &types.AttributeValueMemberS{
			Value: v,
		}
	case json.Number:
		return &types.AttributeValueMemberS{
			Value: v.String(),
		}
	case float64:
		return &types.AttributeValueMemberN{
			Value: formatFloat(v),
		}
	case int:
		return &types.AttributeValueMemberN{
			Value: strconv.Itoa(v),
		}
	case int64:
		return &types.AttributeValueMemberN{
			Value: strconv.Itoa(int(v)),
		}
	case bool:
		return &types.AttributeValueMemberBOOL{
			Value: v,
		}
	case nil:
		return &types.AttributeValueMemberNULL{
			Value: true,
		}
	}
	return &types.AttributeValueMemberS{
		Value: fmt.Sprintf("%v", root),
	}
}

func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 4, 64)
	s = strings.TrimRight(s, "0")
	if strings.HasSuffix(s, ".") {
		s += "0"
	}
	return s
}

func MarshalJSONToDynamoDBAttribute(key, path string, root any, keyTypeMap map[string]neosync_types.KeyType) types.AttributeValue {
	gObj := gabs.Wrap(root)
	if path != "" {
		gObj = gObj.Path(path)
	}
	return MarshalToAttributeValue(key, gObj.Data(), keyTypeMap)
}

func ConvertToStringSlice(slice any) ([]string, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("input is not a slice")
	}

	result := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i).Interface()
		result[i] = anyToString(elem)
	}

	return result, nil
}

func anyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case float32, float64:
		return formatFloat(reflect.ValueOf(v).Float())
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func ConvertAttributeValueToGoMap(av dynamotypes.AttributeValue) (map[string]any, error) {
	switch v := av.(type) {
	case *dynamotypes.AttributeValueMemberS:
		return map[string]any{"S": v.Value}, nil
	case *dynamotypes.AttributeValueMemberB:
		return map[string]any{"B": v.Value}, nil
	case *dynamotypes.AttributeValueMemberN:
		return map[string]any{"N": v.Value}, nil
	case *dynamotypes.AttributeValueMemberBOOL:
		return map[string]any{"BOOL": v.Value}, nil
	case *dynamotypes.AttributeValueMemberNULL:
		return map[string]any{"NULL": v.Value}, nil
	case *dynamotypes.AttributeValueMemberM:
		m := make(map[string]any)
		for k, val := range v.Value {
			var err error
			m[k], err = ConvertAttributeValueToGoMap(val)
			if err != nil {
				return nil, err
			}
		}
		return map[string]any{"M": m}, nil
	case *dynamotypes.AttributeValueMemberL:
		l := make([]any, len(v.Value))
		for i, val := range v.Value {
			var err error
			l[i], err = ConvertAttributeValueToGoMap(val)
			if err != nil {
				return nil, err
			}
		}
		return map[string]any{"L": l}, nil
	case *dynamotypes.AttributeValueMemberSS:
		return map[string]any{"SS": v.Value}, nil
	case *dynamotypes.AttributeValueMemberNS:
		return map[string]any{"NS": v.Value}, nil
	case *dynamotypes.AttributeValueMemberBS:
		return map[string]any{"BS": v.Value}, nil
	default:
		return nil, fmt.Errorf("unsupported AttributeValue type")
	}
}

func ConvertDynamoItemToGoMap(input map[string]dynamotypes.AttributeValue) (map[string]any, error) {
	result := make(map[string]any)

	for key, av := range input {
		dynamoDBJSON, err := ConvertAttributeValueToGoMap(av)
		if err != nil {
			return nil, fmt.Errorf("error converting key %s: %w", key, err)
		}
		result[key] = dynamoDBJSON
	}

	return result, nil
}
