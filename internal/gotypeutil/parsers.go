package gotypeutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
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

func MapToJson(m any) ([]byte, error) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error marshaling map to JSON: %w", err)
	}
	return jsonData, nil
}

func JsonToMap(j []byte) (map[string]any, error) {
	var jMap map[string]any
	err := json.Unmarshal(j, &jMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}
	return jMap, nil
}

func ParseSlice(input any) ([]any, error) {
	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("value is not a slice")
	}

	length := v.Len()
	result := make([]any, length)
	for i := 0; i < length; i++ {
		result[i] = v.Index(i).Interface()
	}
	return result, nil
}

func ToPtr[T any](value T) *T {
	return &value
}
