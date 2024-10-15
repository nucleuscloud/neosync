package transformer_utils

import "reflect"

func IsZeroValue[T any](value T) bool {
	var zero T

	if reflect.DeepEqual(value, zero) {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	}

	return false
}
