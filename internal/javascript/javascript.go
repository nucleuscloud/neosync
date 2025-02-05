package javascript

import (
	"log/slog"

	goja_require "github.com/dop251/goja_nodejs/require"
	javascript_functions "github.com/nucleuscloud/neosync/internal/javascript/functions"
	benthos_functions "github.com/nucleuscloud/neosync/internal/javascript/functions/benthos"
	neosync_functions "github.com/nucleuscloud/neosync/internal/javascript/functions/neosync"
	javascript_vm "github.com/nucleuscloud/neosync/internal/javascript/vm"
)

// Comes full featured, but expects a value api that the benthos/neosync functions can manipulate
func NewDefaultValueRunner(
	valueApi javascript_functions.ValueApi,
	logger *slog.Logger,
) (*javascript_vm.Runner, error) {
	functions, err := getDefaultFunctions()
	if err != nil {
		return nil, err
	}
	return javascript_vm.NewRunner(
		javascript_vm.WithValueApi(valueApi),
		javascript_vm.WithLogger(logger),
		javascript_vm.WithConsole(),
		javascript_vm.WithJsRegistry(goja_require.NewRegistry()),
		javascript_vm.WithFunctions(functions...),
	)
}

// Comes full featured but does not register any custom functions
func NewDefaultRunner(
	logger *slog.Logger,
) (*javascript_vm.Runner, error) {
	return javascript_vm.NewRunner(
		javascript_vm.WithLogger(logger),
		javascript_vm.WithConsole(),
		javascript_vm.WithJsRegistry(goja_require.NewRegistry()),
	)
}

func getDefaultFunctions() ([]*javascript_functions.FunctionDefinition, error) {
	benthosFns := benthos_functions.Get()
	neosyncFns, err := neosync_functions.Get()
	if err != nil {
		return nil, err
	}
	output := make([]*javascript_functions.FunctionDefinition, 0, len(benthosFns)+len(neosyncFns))
	output = append(output, benthosFns...)
	output = append(output, neosyncFns...)
	return output, nil
}
