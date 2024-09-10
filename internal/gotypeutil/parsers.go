package gotypeutil

import (
	"encoding/json"
	"errors"
	"fmt"
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

func MapToJson(m map[string]any) ([]byte, error) {
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

func ParseSliceToMapSlice(input []any) ([]map[string]any, error) {
	result := make([]map[string]any, 0, len(input))

	for _, item := range input {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		} else {
			return nil, fmt.Errorf("item is not of type map[string]any")
		}
	}

	return result, nil
}
