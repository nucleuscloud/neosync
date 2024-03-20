package neosync_benthos_error

import (
	"context"

	"github.com/benthosdev/benthos/v4/public/service"
)

func errorProcessorSpec() *service.ConfigSpec {
	return service.NewConfigSpec()
}

// Registers an processor on a benthos environment called error
func RegisterErrorProcessor(env *service.Environment, stopWorkflowChannel chan bool) error {
	return env.RegisterBatchProcessor(
		"error", errorProcessorSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			return newErrorProcessor(mgr.Logger(), stopWorkflowChannel), nil
		})
}

type errorProcessor struct {
	logger              *service.Logger
	stopWorkflowChannel chan bool
}

func newErrorProcessor(logger *service.Logger, channel chan bool) *errorProcessor {
	return &errorProcessor{
		logger:              logger,
		stopWorkflowChannel: channel,
	}
}

func (r *errorProcessor) ProcessBatch(context.Context, service.MessageBatch) ([]service.MessageBatch, error) {
	r.logger.Infof("Error processor: sending stop workflow signal")
	r.stopWorkflowChannel <- true
	return []service.MessageBatch{}, nil
}

func (r *errorProcessor) Close(ctx context.Context) error {
	return nil
}
