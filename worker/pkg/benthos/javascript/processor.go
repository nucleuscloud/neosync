package javascript_processor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/dop251/goja"
	"github.com/nucleuscloud/neosync/internal/benthos_slogger"
	"github.com/nucleuscloud/neosync/internal/javascript"
	javascript_vm "github.com/nucleuscloud/neosync/internal/javascript/vm"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"

	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	codeField = "code"
)

func javascriptProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewInterpolatedStringField(codeField))
}

func RegisterNeosyncJavascriptProcessor(
	env *service.Environment,
	transformPiiTextApi transformers.TransformPiiTextApi,
) error {
	return env.RegisterBatchProcessor(
		"neosync_javascript", javascriptProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			return newJavascriptProcessorFromConfig(conf, mgr, transformPiiTextApi)
		})
}

type javascriptProcessor struct {
	program *goja.Program
	slogger *slog.Logger
	vmPool  sync.Pool
}

func newJavascriptProcessorFromConfig(
	conf *service.ParsedConfig,
	mgr *service.Resources,
	transformPiiTextApi transformers.TransformPiiTextApi,
) (*javascriptProcessor, error) {
	code, err := conf.FieldString(codeField)
	if err != nil {
		return nil, err
	}

	filename := "main.js"
	program, err := goja.Compile(filename, code, false)
	if err != nil {
		return nil, fmt.Errorf("failed to compile javascript code: %v", err)
	}

	logger := mgr.Logger()
	slogger := benthos_slogger.NewSlogger(logger)

	return &javascriptProcessor{
		program: program,
		slogger: slogger,
		vmPool: sync.Pool{
			New: func() any {
				val, err := newPoolItem(slogger, transformPiiTextApi)
				if err != nil {
					return err
				}
				return val
			},
		},
	}, nil
}

type vmPoolItem struct {
	runner   *javascript_vm.Runner
	valueApi *benthosValueApi
}

func (j *javascriptProcessor) ProcessBatch(
	ctx context.Context,
	batch service.MessageBatch,
) (result []service.MessageBatch, err error) {
	var runner *javascript_vm.Runner
	var valueApi *benthosValueApi

	switch poolItem := j.vmPool.Get().(type) {
	case *vmPoolItem:
		runner = poolItem.runner
		valueApi = poolItem.valueApi
		defer func() {
			poolItem.valueApi.SetMessage(nil) // reset the message to nil
			j.vmPool.Put(poolItem)
		}()
	case error:
		return nil, poolItem
	}

	// Add panic recovery for the entire batch processing
	// Goja has panic recovery built in, but if it encounters an uncatchable panic
	// An uncatchable panic is one that happens in a Go function called by JS.
	// Goja looks for a special `uncatchableException` error type and traps that in the panic.
	// For anything else though, it will re-panic with the original paniced error. /facepalm
	// This here acts as a final catch-all defense for anything we missed so prevent the process from crashing.
	defer func() {
		if r := recover(); r != nil {
			j.slogger.Error(
				"recovered from panic in neosync_javascript batch processor",
				"error",
				fmt.Sprintf("%v", r),
			)
			// Set the named return value 'err'
			err = fmt.Errorf("neosync_javascript batch processor panic recovered: %v", r)
			return
		}
	}()

	var newBatch service.MessageBatch

	for i := range batch {
		valueApi.SetMessage(batch[i])
		_, err := runner.Run(ctx, j.program)
		if err != nil {
			return nil, err
		}
		if newMsg := valueApi.Message(); newMsg != nil {
			newBatch = append(newBatch, newMsg)
		}
	}

	return []service.MessageBatch{newBatch}, nil
}

func (j *javascriptProcessor) Close(ctx context.Context) error {
	return nil
}

func newPoolItem(
	logger *slog.Logger,
	transformPiiTextApi transformers.TransformPiiTextApi,
) (*vmPoolItem, error) {
	valueApi := newBatchBenthosValueApi()
	runner, err := javascript.NewDefaultValueRunner(valueApi, transformPiiTextApi, logger)
	if err != nil {
		return nil, err
	}
	return &vmPoolItem{
		valueApi: valueApi,
		runner:   runner,
	}, nil
}
