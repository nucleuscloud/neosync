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

	"github.com/warpstreamlabs/bento/public/service"
)

const (
	codeField = "code"
)

func javascriptProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewInterpolatedStringField(codeField))
}

func RegisterNeosyncJavascriptProcessor(env *service.Environment) error {
	return env.RegisterBatchProcessor(
		"neosync_javascript", javascriptProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			return newJavascriptProcessorFromConfig(conf, mgr)
		})
}

type javascriptProcessor struct {
	program *goja.Program
	slogger *slog.Logger
	vmPool  sync.Pool
}

func newJavascriptProcessorFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*javascriptProcessor, error) {
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
				val, err := newPoolItem(slogger)
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

func (j *javascriptProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
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

func newPoolItem(logger *slog.Logger) (*vmPoolItem, error) {
	valueApi := newBatchBenthosValueApi()
	runner, err := javascript.NewDefaultValueRunner(valueApi, logger)
	if err != nil {
		return nil, err
	}
	return &vmPoolItem{
		valueApi: valueApi,
		runner:   runner,
	}, nil
}
