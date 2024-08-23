package mongomanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MetaTypeMapStr = "neosync_key_type_map"
)

type KeyType int

const (
	StringSet KeyType = iota
	NumberSet
	ObjectID
	Decimal128
)

func ParseStringAsNumber(s string) (any, error) {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i, nil
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}

	return nil, errors.New("input string is neither a valid int nor a float")
}

func UnmarshalBSONDocument(doc bson.D) (standardMap map[string]any, keyTypeMap map[string]KeyType) {
	result := make(map[string]any)
	ktm := make(map[string]KeyType)
	for _, elem := range doc {
		result[elem.Key] = ParseBSONValue(elem.Key, elem.Value, ktm)
	}
	return result, ktm
}

func ParseBSONValue(key string, value any, keyTypeMap map[string]KeyType) any {
	fmt.Println(key, value)
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		return v
	case int32, int64, float64:
		return v
	case bool:
		return v
	case primitive.Decimal128:
		keyTypeMap[key] = Decimal128
		return v.String()
	case primitive.Binary:
		return map[string]any{
			"base64":  v.Data,
			"subType": v.Subtype,
		}
	case primitive.ObjectID:
		keyTypeMap[key] = ObjectID
		return v.Hex()
	case primitive.DateTime:
		return v.Time()
	case primitive.Timestamp:
		return map[string]any{
			"t": v.T,
			"i": v.I,
		}
	case primitive.Null:
		return nil
	case primitive.Undefined:
		return nil
	case primitive.Regex:
		return map[string]any{
			"pattern": v.Pattern,
			"options": v.Options,
		}
	case bson.D:
		m := make(map[string]any)
		for _, elem := range v {
			path := elem.Key
			if key != "" {
				path = fmt.Sprintf("%s.%s", key, elem.Key)
			}
			m[elem.Key] = ParseBSONValue(path, elem.Value, keyTypeMap)
		}
		return m
	case bson.A:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = ParseBSONValue(fmt.Sprintf("%s[%d]", key, i), item, keyTypeMap)
		}
		return result
	default:
		return fmt.Sprintf("%v", v)
	}
}

func MarshalToBSONValue(key string, root any, keyTypeMap map[string]KeyType) any {
	fmt.Printf("key: %s  Type of root: %v\n", key, reflect.TypeOf(root))

	if root == nil {
		return nil
	}

	// Handle pointers
	val := reflect.ValueOf(root)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	root = val.Interface()
	if typeStr, ok := keyTypeMap[key]; ok {
		switch typeStr {
		case Decimal128:
			// return primitive.ParseDecimal128(root)
			vStr, ok := root.(string)
			if ok {
				d, err := primitive.ParseDecimal128(vStr)
				if err == nil {
					fmt.Println("returning decimal128")
					return d
				}
			}
			fmt.Println("returning decimal as string")
			return fmt.Sprintf("%v", root)

		case ObjectID:
			vStr, ok := root.(string)
			if ok {
				if key == "_id" || key == "$set._id" {
					return vStr
				}
				if objectID, err := primitive.ObjectIDFromHex(vStr); err == nil {
					fmt.Println("object id", key)
					return objectID
				}
			}
			return fmt.Sprintf("%v", root)
			// case DateTime:
			// 	return root
		}
	}

	switch v := root.(type) {
	case map[string]any:
		// Handle Binary
		if base64Data, ok := v["base64"].([]byte); ok {
			if subType, ok := v["subType"].(uint8); ok {
				return primitive.Binary{Data: base64Data, Subtype: subType}
			}
		}

		// Handle Timestamp
		if t, ok := v["t"].(uint32); ok {
			if i, ok := v["i"].(uint32); ok {
				return primitive.Timestamp{T: t, I: i}
			}
		}

		// Handle Regex
		if pattern, ok := v["pattern"].(string); ok {
			if options, ok := v["options"].(string); ok {
				return primitive.Regex{Pattern: pattern, Options: options}
			}
		}

		// Default map handling
		doc := bson.D{}
		for k, v2 := range v {
			path := k
			if key != "" {
				path = fmt.Sprintf("%s.%s", key, k)
			}
			doc = append(doc, bson.E{Key: k, Value: MarshalToBSONValue(path, v2, keyTypeMap)})
		}
		return doc

	case []byte:
		return primitive.Binary{Data: v}

	case []any:
		a := bson.A{}
		for i, v2 := range v {
			a = append(a, MarshalToBSONValue(fmt.Sprintf("%s[%d]", key, i), v2, keyTypeMap))
		}
		return a

	case string:
		return v

	case json.Number:
		n, err := v.Int64()
		if err == nil {
			return n
		}
		f, err := v.Float64()
		if err == nil {
			return f
		}
		return v.String()

	case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return v

	case bool:
		return v

	case time.Time:
		return primitive.NewDateTimeFromTime(v)

	case nil:
		return primitive.Null{}

	default:
		return fmt.Sprintf("%v", v)
	}
}

func MarshalJSONToBSONDocument(root any, keyTypeMap map[string]KeyType) bson.D {
	m, ok := root.(map[string]any)
	if !ok {
		return bson.D{}
	}

	doc := bson.D{}
	for k, v := range m {
		doc = append(doc, bson.E{Key: k, Value: MarshalToBSONValue(k, v, keyTypeMap)})
	}
	return doc
}

func ConvertToSlice(slice any) ([]any, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("input is not a slice")
	}

	result := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}

	return result, nil
}

// New function to convert a Timestamp to time.Time
func TimestampToTime(ts primitive.Timestamp) time.Time {
	return time.Unix(int64(ts.T), 0)
}
