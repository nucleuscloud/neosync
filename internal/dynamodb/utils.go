package dynamodb

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	MetaTypeMapStr = "neosync_key_type_map"
)

type KeyType int

const (
	StringSet KeyType = iota
	NumberSet
)

func ConvertStringToNumber(s string) (any, error) {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i, nil
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}

	return nil, errors.New("input string is neither a valid int nor a float")
}

func ConvertDynamoDBItemToMap(item map[string]any) (standardMap map[string]any, keyTypeMap map[string]KeyType) {
	result := make(map[string]any)
	ktm := make(map[string]KeyType)
	for key, value := range item {
		result[key] = ConvertDynamoDBValue(key, value, ktm)
	}

	return result, ktm
}

func ConvertDynamoDBValue(key string, value any, keyTypeMap map[string]KeyType) any {
	if m, ok := value.(map[string]any); ok {
		for dynamoType, dynamoValue := range m {
			switch dynamoType {
			case "S":
				return dynamoValue.(string)
			case "B":
				s := dynamoValue.(string)
				byteSlice, err := base64.StdEncoding.DecodeString(s)
				if err != nil {
					return dynamoValue
				}
				return byteSlice
			case "N":
				n, err := ConvertStringToNumber(dynamoValue.(string))
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
					result[i] = ConvertDynamoDBValue(fmt.Sprintf("%s[%d]", key, i), item, keyTypeMap)
				}
				return result
			case "M":
				mAny := map[string]any{}
				for k, v := range dynamoValue.(map[string]any) {
					path := k
					if key != "" {
						path = fmt.Sprintf("%s.%s", key, k)
					}
					val := ConvertDynamoDBValue(path, v, keyTypeMap)
					mAny[k] = val
				}
				return mAny
			case "BS":
				bytes := dynamoValue.([]any)
				result := make([][]byte, len(bytes))
				for i, b := range bytes {
					s := b.(string)
					byteSlice, err := base64.StdEncoding.DecodeString(s)
					if err != nil {
						return dynamoValue
					}

					result[i] = byteSlice
				}
				return result
			case "SS":
				keyTypeMap[key] = StringSet
				ss := dynamoValue.([]any)
				result := make([]string, len(ss))
				for i, s := range ss {
					result[i] = s.(string)
				}
				return result
			case "NS":
				keyTypeMap[key] = NumberSet
				numbers := dynamoValue.([]any)
				result := make([]any, len(numbers))
				for i, num := range numbers {
					n, err := ConvertStringToNumber(num.(string))
					if err != nil {
						result[i] = num
					}
					result[i] = n
				}
				return result
			}
		}
	}
	return value
}

func AttributeValueMapToStandardJSON(item map[string]types.AttributeValue) (standardMap map[string]any, keyTypeMap map[string]KeyType) {
	standardJSON := make(map[string]any)
	ktm := make(map[string]KeyType)
	for k, v := range item {
		val := AttributeValueToStandardValue(k, v, ktm)
		standardJSON[k] = val
	}
	return standardJSON, ktm
}

// attributeValueToStandardValue converts a DynamoDB AttributeValue to a standard value
func AttributeValueToStandardValue(key string, v types.AttributeValue, keyTypeMap map[string]KeyType) any {
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
			val := AttributeValueToStandardValue(fmt.Sprintf("%s[%d]", key, i), v, keyTypeMap)
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
			val := AttributeValueToStandardValue(path, v, keyTypeMap)
			mAny[k] = val
		}
		return mAny
	case *types.AttributeValueMemberN:
		n, err := ConvertStringToNumber(t.Value)
		if err != nil {
			return t.Value
		}
		return n
	case *types.AttributeValueMemberNS:
		keyTypeMap[key] = NumberSet
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			n, err := ConvertStringToNumber(v)
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
		keyTypeMap[key] = StringSet
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			lAny[i] = v
		}
		return lAny
	}
	return nil
}

func AnyToAttributeValue(key string, root any, keyTypeMap map[string]KeyType) types.AttributeValue {
	if typeStr, ok := keyTypeMap[key]; ok {
		switch typeStr {
		case StringSet:
			s, ok := getGenericSlice[string](root)
			if ok {
				return &types.AttributeValueMemberSS{
					Value: s,
				}
			}
		case NumberSet:
			stringSlice, err := toStringSlice(root)
			if err == nil {
				return &types.AttributeValueMemberNS{
					Value: stringSlice,
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
			m[k] = AnyToAttributeValue(path, v2, keyTypeMap)
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
			l[i] = AnyToAttributeValue(fmt.Sprintf("%s[%d]", key, i), v2, keyTypeMap)
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

func getGenericSlice[T any](v any) ([]T, bool) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Slice {
		return nil, false
	}

	genericSlice := make([]T, val.Len())
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i).Interface()
		if tElem, ok := elem.(T); ok {
			genericSlice[i] = tElem
		} else {
			return nil, false
		}
	}

	return genericSlice, true
}

func jsonToMap(key, path string, root any, keyTypeMap map[string]KeyType) types.AttributeValue {
	gObj := gabs.Wrap(root)
	if path != "" {
		gObj = gObj.Path(path)
	}
	return AnyToAttributeValue(key, gObj.Data(), keyTypeMap)
}
