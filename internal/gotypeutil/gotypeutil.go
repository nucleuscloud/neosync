package gotypeutil

import (
	"reflect"
)

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

func IsSlice(val any) bool {
	v := reflect.ValueOf(val)
	return v.Kind() == reflect.Slice
}

func IsMap(val any) bool {
	v := reflect.ValueOf(val)
	return v.Kind() == reflect.Map
}

func IsSliceOfMaps(val any) bool {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Slice {
		return false
	}

	if v.Len() == 0 {
		return false
	}

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Map {
			return false
		}
	}

	return true
}
