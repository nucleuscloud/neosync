package neosync_benthos_error

import (
	"context"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/service"
)

func errorProcessorSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Summary(`Sends stop workflow signal`)
}

// Registers an processor on a benthos environment called error
func RegisterErrorProcessor(env *service.Environment, stopActivityChannel chan bool) error {
	return env.RegisterBatchProcessor(
		"error", errorProcessorSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			return newErrorProcessor(mgr.Logger(), stopActivityChannel), nil
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

func (r *errorProcessor) ProcessBatch(_ context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
	r.logger.Info("Error processor: sending stop workflow signal")
	r.stopWorkflowChannel <- true
	return nil, fmt.Errorf("Processor error occurred. Stopping workflow.")
}

func (r *errorProcessor) Close(ctx context.Context) error {
	return nil
}
