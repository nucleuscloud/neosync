package javascript_vm

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/dop251/goja"
	goja_require "github.com/dop251/goja_nodejs/require"
	javascript_functions "github.com/nucleuscloud/neosync/internal/javascript/functions"
	"github.com/nucleuscloud/neosync/internal/testutil"

	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		runner, err := NewRunner()
		require.NoError(t, err)

		program := goja.MustCompile("test.js", "1+1", true)
		result, err := runner.Run(context.Background(), program)
		require.NoError(t, err)
		require.Equal(t, int64(2), result.ToInteger())
	})

	t.Run("with_console", func(t *testing.T) {
		runner, err := NewRunner(WithConsole(), WithJsRegistry(goja_require.NewRegistry()))
		require.NoError(t, err)

		program := goja.MustCompile("test.js", "console.log('hello world')", true)
		_, err = runner.Run(context.Background(), program)
		require.NoError(t, err)
	})

	t.Run("with_console_and_logger", func(t *testing.T) {
		runner, err := NewRunner(WithConsole(), WithJsRegistry(goja_require.NewRegistry()), WithLogger(testutil.GetTestLogger(t)))
		require.NoError(t, err)

		program := goja.MustCompile("test.js", `console.log('hello world');`, true)
		_, err = runner.Run(context.Background(), program)
		require.NoError(t, err)
	})

	t.Run("parallel_runs", func(t *testing.T) {
		runner, err := NewRunner(WithConsole(), WithJsRegistry(goja_require.NewRegistry()), WithLogger(testutil.GetTestLogger(t)))
		require.NoError(t, err)

		program := goja.MustCompile("test.js", `console.log('hello world');`, true)
		wg := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err = runner.Run(context.Background(), program)
				require.NoError(t, err)
			}()
		}
		wg.Wait()
	})

	t.Run("with_functions", func(t *testing.T) {
		customFn := javascript_functions.NewFunctionDefinition("test", "test", func(r javascript_functions.Runner) javascript_functions.Function {
			return func(call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
				return "hello world", nil
			}
		})

		runner, err := NewRunner(WithFunctions(customFn))
		require.NoError(t, err)

		program := goja.MustCompile("test.js", `test.test();`, true)
		result, err := runner.Run(context.Background(), program)
		require.NoError(t, err)
		require.Equal(t, "hello world", result.String())
	})
}

func BenchmarkRunner_Single(b *testing.B) {
	runner, err := NewRunner(WithConsole(), WithJsRegistry(goja_require.NewRegistry()), WithLogger(testutil.GetTestLogger(b)))
	require.NoError(b, err)

	program := goja.MustCompile("test.js", `console.log('hello world');`, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = runner.Run(context.Background(), program)
		require.NoError(b, err)
	}
}
