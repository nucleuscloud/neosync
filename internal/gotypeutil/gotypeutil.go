package gotypeutil

import "reflect"

func IsMultiDimensionalSlice(val any) bool {
	rv := reflect.ValueOf(val)

	if rv.Kind() != reflect.Slice {
		return false
	}

	// if the slice is empty can't determine if it's multi-dimensional
	if rv.Len() == 0 {
		return false
	}

	firstElem := rv.Index(0)
	// if an interface check underlying value
	if firstElem.Kind() == reflect.Interface {
		firstElem = firstElem.Elem()
	}

	return firstElem.Kind() == reflect.Slice
}
