package javascript_vm

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	javascript_functions "github.com/nucleuscloud/neosync/internal/javascript/functions"
)

type Runner struct {
	vm      *goja.Runtime
	options Options
	mu      sync.Mutex
}

func (r *Runner) ValueApi() javascript_functions.ValueApi {
	return r.options.valueApi
}

type Options struct {
	logger          *slog.Logger
	requireRegistry *require.Registry
	functions       []*javascript_functions.FunctionDefinition
	consoleEnabled  bool
	valueApi        javascript_functions.ValueApi
}

type Option func(*Options)

func WithValueApi(valueApi javascript_functions.ValueApi) Option {
	return func(opts *Options) {
		opts.valueApi = valueApi
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(opts *Options) {
		opts.logger = logger
	}
}

func WithJsRegistry(registry *require.Registry) Option {
	return func(opts *Options) {
		opts.requireRegistry = registry
	}
}

func WithFunctions(functions ...*javascript_functions.FunctionDefinition) Option {
	return func(opts *Options) {
		opts.functions = functions
	}
}

func WithConsole() Option {
	return func(opts *Options) {
		opts.consoleEnabled = true
	}
}

// Creates a new JS Runner
func NewRunner(opts ...Option) (*Runner, error) {
	options := Options{logger: slog.Default()}
	for _, opt := range opts {
		opt(&options)
	}

	vm := goja.New()

	// if the stars align, we'll register the custom console module with the logger
	// must come before requireRegistry.Enable()
	if options.requireRegistry != nil && options.consoleEnabled && options.logger != nil {
		options.requireRegistry.RegisterNativeModule(console.ModuleName, console.RequireWithPrinter(newConsoleLogger(stdPrefix, options.logger)))
	}

	if options.requireRegistry != nil {
		options.requireRegistry.Enable(vm)
	}

	// must come after requireRegistry.Enable()
	if options.consoleEnabled {
		console.Enable(vm)
	}

	runner := &Runner{
		vm:      vm,
		options: options,
	}

	for _, function := range options.functions {
		if err := registerFunction(runner, function); err != nil {
			return nil, err
		}
	}

	return runner, nil
}

func (r *Runner) Run(ctx context.Context, program *goja.Program) (goja.Value, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.vm.RunProgram(program)
}

// Registers a custom function with the vm
func registerFunction(runner *Runner, function *javascript_functions.FunctionDefinition) error {
	var targetObj *goja.Object
	if targetObjValue := runner.vm.GlobalObject().Get(function.Namespace()); targetObjValue != nil {
		targetObj = targetObjValue.ToObject(runner.vm)
	}
	if targetObj == nil {
		if err := runner.vm.GlobalObject().Set(function.Namespace(), map[string]any{}); err != nil {
			return fmt.Errorf("failed to set global %s object: %w", function.Namespace(), err)
		}
		targetObj = runner.vm.GlobalObject().Get(function.Namespace()).ToObject(runner.vm)
	}

	if err := targetObj.Set(function.Name(), func(call goja.FunctionCall, rt *goja.Runtime) goja.Value {
		l := runner.options.logger.With("function", function.Name())
		fn := function.Ctor()(runner)
		result, err := fn(context.Background(), call, rt, l)
		if err != nil {
			// This _has_ to be a panic so that the error is properly thrown in the JS runtime
			// Otherwise things like try/catch will not work properly
			panic(rt.ToValue(err.Error()))
		}
		return rt.ToValue(result)
	}); err != nil {
		return fmt.Errorf("failed to set global %s function %v: %w", function.Namespace(), function.Name(), err)
	}
	return nil
}
