package javascript_functions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/dop251/goja"
)

type Runner interface {
	ValueApi() ValueApi
}

type ValueApi interface {
	SetBytes(value []byte)
	AsBytes() ([]byte, error)
	SetStructured(value any)
	AsStructured() (any, error)

	MetaGet(name string) (any, bool)
	MetaSetMut(name string, value any)
}

type Ctor func(r Runner) Function

type FunctionDefinition struct {
	namespace string
	name      string
	// ctor means "constructor"
	ctor Ctor
}

func NewFunctionDefinition(namespace, name string, ctor Ctor) *FunctionDefinition {
	return &FunctionDefinition{
		namespace: namespace,
		name:      name,
		ctor:      ctor,
	}
}

func (f *FunctionDefinition) Namespace() string {
	return f.namespace
}

func (f *FunctionDefinition) Name() string {
	return f.name
}

func (f *FunctionDefinition) Ctor() Ctor {
	return f.ctor
}

type Function func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error)

// getTypeString returns a string representation of the type of a goja.Value
// Returns "null" for nil or null values, and "undefined" for undefined values
func getTypeString(arg goja.Value) string {
	if arg == nil || goja.IsNull(arg) {
		return "null"
	}
	if goja.IsUndefined(arg) {
		return "undefined"
	}
	return arg.ExportType().String()
}

// Takes in a goja function call and returns the parsed arguments into the provided pointers.
// Returns an error if the arguments are not of the expected type.
func ParseFunctionArguments(call goja.FunctionCall, ptrs ...any) error {
	if len(ptrs) < len(call.Arguments) {
		return fmt.Errorf("have %d arguments, but only %d pointers to parse into", len(call.Arguments), len(ptrs))
	}

	for i := range call.Arguments {
		arg, ptr := call.Argument(i), ptrs[i]

		if goja.IsUndefined(arg) {
			return fmt.Errorf("argument at position %d is undefined", i)
		}

		var err error
		switch p := ptr.(type) {
		case *string:
			*p = arg.String()
		case *int:
			*p = int(arg.ToInteger())
		case *int64:
			*p = arg.ToInteger()
		case *float64:
			*p = arg.ToFloat()
		case *map[string]any:
			*p, err = getMapFromValue(arg)
		case *bool:
			*p = arg.ToBoolean()
		case *[]any:
			*p, err = getSliceFromValue(arg)
		case *[]map[string]any:
			*p, err = getMapSliceFromValue(arg)
		case *goja.Value:
			*p = arg
		case *any:
			*p = arg.Export()
		default:
			typeStr := getTypeString(arg)
			return fmt.Errorf("encountered unhandled type %s while trying to parse %v into %v", typeStr, arg, ptr)
		}
		if err != nil {
			typeStr := getTypeString(arg)
			return fmt.Errorf("could not parse %v (%s) into %v (%T): %v", arg, typeStr, ptr, ptr, err)
		}
	}

	return nil
}

func getMapFromValue(val goja.Value) (map[string]any, error) {
	outVal := val.Export()
	if outVal == nil {
		return map[string]any{}, nil
	}
	v, ok := outVal.(map[string]any)
	if !ok {
		return nil, errors.New("value is not of type map")
	}
	return v, nil
}

func getSliceFromValue(val goja.Value) ([]any, error) {
	outVal := val.Export()
	if outVal == nil {
		return []any{}, nil
	}
	v, ok := outVal.([]any)
	if !ok {
		return nil, errors.New("value is not of type slice")
	}
	return v, nil
}

func getMapSliceFromValue(val goja.Value) ([]map[string]any, error) {
	outVal := val.Export()
	if outVal == nil {
		return []map[string]any{}, nil
	}
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
