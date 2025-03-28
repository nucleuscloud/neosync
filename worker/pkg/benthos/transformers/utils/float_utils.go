package transformer_utils

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
)

/* FLOAT MANIPULATION UTILS */

// Generates a random float64 in the range of the min and max float64 values
func GenerateRandomFloat64WithInclusiveBounds(
	randomizer rng.Rand,
	minValue, maxValue float64,
) (float64, error) {
	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}

	if minValue == maxValue {
		return minValue, nil
	}

	randValue := randomizer.Float64()

	// Scale and shift the value to the range
	returnValue := minValue + randValue*(maxValue-minValue)
	return returnValue, nil
}

func AnyToFloat64(val any) (float64, error) {
	if val == nil {
		return 0, errors.New("value must not be nil")
	}
	switch t := val.(type) {
	case string:
		return strconv.ParseFloat(t, 64)
	case []byte:
		return strconv.ParseFloat(string(t), 64)
	case int:
		return float64(t), nil
	case *int:
		return float64(*t), nil
	case int8:
		return float64(t), nil
	case *int8:
		return float64(*t), nil
	case int16:
		return float64(t), nil
	case *int16:
		return float64(*t), nil
	case int32:
		return float64(t), nil
	case *int32:
		return float64(*t), nil
	case int64:
		return float64(t), nil
	case *int64:
		return float64(*t), nil
	case uint:
		return float64(t), nil
	case *uint:
		return float64(*t), nil
	case uint8:
		return float64(t), nil
	case *uint8:
		return float64(*t), nil
	case uint16:
		return float64(t), nil
	case *uint16:
		return float64(*t), nil
	case uint32:
		return float64(t), nil
	case *uint32:
		return float64(*t), nil
	case uint64:
		return float64(t), nil
	case *uint64:
		return float64(*t), nil
	case float32:
		return float64(t), nil
	case *float32:
		return float64(*t), nil
	case float64:
		return t, nil
	case *float64:
		return *t, nil
	case *big.Float:
		float64Val, _ := t.Float64()
		return float64Val, nil
	case big.Float:
		float64Val, _ := t.Float64()
		return float64Val, nil
	case bool:
		if t {
			return 1, nil
		}
		return 0, nil
	case *bool:
		if *t {
			return 1, nil
		}
		return 0, nil
	default:
		return -1, fmt.Errorf("converting type %T to float64 is not currently supported", t)
	}
}
