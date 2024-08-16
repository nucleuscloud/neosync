package neosync_benthos_dynamodb

import (
	"errors"
	"strconv"
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
