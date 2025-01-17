package javascript_functions

import (
	"log/slog"

	"github.com/dop251/goja"
)

type Runner interface{}

type Ctor func(r Runner) Function

type FunctionDefinition struct {
	namespace string
	name      string
	// params    []*FunctionParam
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

// type FunctionParam struct {
// 	name    string
// 	typeStr string
// 	what    string
// }

type Function func(call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error)
