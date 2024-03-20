package neosync_benthos_error

import (
	"context"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/service"
)

const (
	fieldsMapping = "fields_mapping"
)

func errorProcessorSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Summary(`Sends stop workflow signal`).
		Fields(
			service.NewBloblangField(fieldsMapping),
		)
}

// Registers an processor on a benthos environment called error
func RegisterErrorProcessor(env *service.Environment, stopWorkflowChannel chan bool) error {
	return env.RegisterBatchProcessor(
		"error_stop_workflow", errorProcessorSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			return newErrorProcessor(conf, mgr.Logger(), stopWorkflowChannel)
		})
}

type errorProcessor struct {
	logger              *service.Logger
	stopWorkflowChannel chan bool
	fieldsMapping       *bloblang.Executor
}

func newErrorProcessor(conf *service.ParsedConfig, logger *service.Logger, channel chan bool) (*errorProcessor, error) {
	fieldsMapping, err := conf.FieldBloblang(fieldsMapping)
	if err != nil {
		return nil, err
	}
	return &errorProcessor{
		fieldsMapping:       fieldsMapping,
		logger:              logger,
		stopWorkflowChannel: channel,
	}, nil
}

func (r *errorProcessor) ProcessBatch(context.Context, service.MessageBatch) ([]service.MessageBatch, error) {
	r.logger.Infof("Error processor: sending stop workflow signal: %s")
	r.stopWorkflowChannel <- true
	return []service.MessageBatch{}, nil
}

func (r *errorProcessor) Close(ctx context.Context) error {
	return nil
}
