package javascript

import (
	"errors"

	"github.com/dop251/goja"
)

func getMapFromValue(val goja.Value) (map[string]any, error) {
	outVal := val.Export()
	v, ok := outVal.(map[string]any)
	if !ok {
		return nil, errors.New("value is not of type map")
	}
	return v, nil
}

func getSliceFromValue(val goja.Value) ([]any, error) {
	outVal := val.Export()
	v, ok := outVal.([]any)
	if !ok {
		return nil, errors.New("value is not of type slice")
	}
	return v, nil
}

func getMapSliceFromValue(val goja.Value) ([]map[string]any, error) {
	outVal := val.Export()
	if v, ok := outVal.([]map[string]any); ok {
		return v, nil
	}
	vSlice, ok := outVal.([]any)
	if !ok {
		return nil, errors.New("value is not of type map slice")
	}
	v := make([]map[string]any, len(vSlice))
	for i, e := range vSlice {
		v[i], ok = e.(map[string]any)
		if !ok {
			return nil, errors.New("value is not of type map slice")
		}
	}
	return v, nil
}
