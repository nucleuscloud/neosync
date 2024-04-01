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
func RegisterErrorProcessor(env *service.Environment, stopActivityChannel chan error) error {
	return env.RegisterBatchProcessor(
		"error", errorProcessorSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			return newErrorProcessor(mgr.Logger(), stopActivityChannel), nil
		})
}

type errorProcessor struct {
	logger              *service.Logger
	stopWorkflowChannel chan error
}

func newErrorProcessor(logger *service.Logger, channel chan error) *errorProcessor {
	return &errorProcessor{
		logger:              logger,
		stopWorkflowChannel: channel,
	}
}

func (r *errorProcessor) ProcessBatch(_ context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
	r.logger.Error("Error processor: sending stop activity signal")
	r.stopWorkflowChannel <- fmt.Errorf("Processor Error") // todo replace this with real error
	return []service.MessageBatch{}, nil
}

func (r *errorProcessor) Close(ctx context.Context) error {
	return nil
}
