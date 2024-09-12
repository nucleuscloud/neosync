package mongomanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"time"

	neosync_types "github.com/nucleuscloud/neosync/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UnmarshalPrimitives(doc map[string]any) (standardMap map[string]any, keyTypeMap map[string]neosync_types.KeyType) {
	result := make(map[string]any)
	ktm := make(map[string]neosync_types.KeyType)
	for k, v := range doc {
		result[k] = ParsePrimitives(k, v, ktm)
	}
	return result, ktm
}

func ParsePrimitives(key string, value any, keyTypeMap map[string]neosync_types.KeyType) any {
	switch v := value.(type) {
	case primitive.Decimal128:
		keyTypeMap[key] = neosync_types.Decimal128
		floatVal, _, err := big.ParseFloat(v.String(), 10, 128, big.ToNearestEven)
		if err == nil {
			return floatVal
		}
		return v
	case primitive.Binary:
		keyTypeMap[key] = neosync_types.Binary
		return v
	case primitive.ObjectID:
		keyTypeMap[key] = neosync_types.ObjectID
		return v
	case primitive.Timestamp:
		keyTypeMap[key] = neosync_types.Timestamp
		return v
	case bson.D:
		m := make(map[string]any)
		for _, elem := range v {
			path := elem.Key
			if key != "" {
				path = fmt.Sprintf("%s.%s", key, elem.Key)
			}
			m[elem.Key] = ParsePrimitives(path, elem.Value, keyTypeMap)
		}
		return m
	case bson.A:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = ParsePrimitives(fmt.Sprintf("%s[%d]", key, i), item, keyTypeMap)
		}
		return result
	default:
		return v
	}
}

func MarshalToBSONValue(key string, root any, keyTypeMap map[string]neosync_types.KeyType) any {
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
		case neosync_types.Decimal128:
			_, ok := root.(primitive.Decimal128)
			if ok {
				return root
			}
			vStr, ok := root.(string)
			if ok {
				d, err := primitive.ParseDecimal128(vStr)
				if err == nil {
					return d
				}
			}
			vFloat, ok := root.(float64)
			if ok {
				d, err := primitive.ParseDecimal128(strconv.FormatFloat(vFloat, 'f', 4, 64))
				if err == nil {
					return d
				}
			}
			vBigFloat, ok := root.(big.Float)
			if ok {
				d, err := primitive.ParseDecimal128(vBigFloat.String())
				if err == nil {
					return d
				}
			}
		case neosync_types.Timestamp:
			_, ok := root.(primitive.Timestamp)
			if ok {
				return root
			}
			t, err := toUint32(root)
			if err == nil {
				return primitive.Timestamp{
					T: t,
					I: 1,
				}
			}

		case neosync_types.ObjectID:
			_, ok := root.(primitive.ObjectID)
			if ok {
				return root
			}
			vStr, ok := root.(string)
			if ok {
				if objectID, err := primitive.ObjectIDFromHex(vStr); err == nil {
					return objectID
				}
			}
		}
	}

	switch v := root.(type) {
	case map[string]any:
		doc := bson.D{}
		for k, v2 := range v {
			path := k
			if key != "" {
				path = fmt.Sprintf("%s.%s", key, k)
			}
			if path == "$set._id" {
				// don't set _id
				continue
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

	case time.Time:
		return primitive.NewDateTimeFromTime(v)

	case nil:
		return primitive.Null{}

	default:
		return v
	}
}

func MarshalJSONToBSONDocument(root any, keyTypeMap map[string]neosync_types.KeyType) bson.D {
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

func toUint32(value any) (uint32, error) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return 0, errors.New("cannot convert negative int to uint32")
		}
		if v > math.MaxUint32 {
			return 0, errors.New("int value out of range for uint32")
		}
		return uint32(v), nil
	case int8:
		if v < 0 {
			return 0, errors.New("cannot convert negative int8 to uint32")
		}
		return uint32(v), nil
	case int16:
		if v < 0 {
			return 0, errors.New("cannot convert negative int16 to uint32")
		}
		return uint32(v), nil
	case int32:
		if v < 0 {
			return 0, errors.New("cannot convert negative int32 to uint32")
		}
		return uint32(v), nil
	case int64:
		if v < 0 || v > math.MaxUint32 {
			return 0, errors.New("int64 value out of range for uint32")
		}
		return uint32(v), nil
	case uint:
		if v > math.MaxUint32 {
			return 0, errors.New("uint value out of range for uint32")
		}
		return uint32(v), nil
	case uint8:
		return uint32(v), nil
	case uint16:
		return uint32(v), nil
	case uint32:
		return v, nil
	case uint64:
		if v > math.MaxUint32 {
			return 0, errors.New("uint64 value out of range for uint32")
		}
		return uint32(v), nil
	case float32:
		if v < 0 || v > math.MaxUint32 || float32(uint32(v)) != v {
			return 0, errors.New("float32 value out of range or not representable as uint32")
		}
		return uint32(v), nil
	case float64:
		if v < 0 || v > math.MaxUint32 || float64(uint32(v)) != v {
			return 0, errors.New("float64 value out of range or not representable as uint32")
		}
		return uint32(v), nil
	case string:
		num, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string to uint32: %v", err)
		}
		return uint32(num), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}
