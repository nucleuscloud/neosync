package neosync_benthos_error

import (
	"context"

	"github.com/benthosdev/benthos/v4/public/service"
)

func errorOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Summary(`Sends stop Activity signal`)
}

// Registers an processor on a benthos environment called error
func RegisterErrorOutput(env *service.Environment, stopWorkflowChannel chan bool) error {
	return env.RegisterBatchOutput(
		"error", errorOutputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchOutput, service.BatchPolicy, int, error) {
			batchPolicy, err := conf.FieldBatchPolicy("batching")
			if err != nil {
				return nil, batchPolicy, -1, err
			}

			maxInFlight, err := conf.FieldInt("max_in_flight")
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			out := newErrorOutput(mgr.Logger(), stopWorkflowChannel)
			return out, batchPolicy, maxInFlight, nil
		})
}

func newErrorOutput(logger *service.Logger, channel chan bool) *errorOutput {
	return &errorOutput{
		logger:              logger,
		stopActivityChannel: channel,
	}
}

type errorOutput struct {
	logger              *service.Logger
	stopActivityChannel chan bool
}

func (e *errorOutput) Connect(ctx context.Context) error {
	return nil
}

func (e *errorOutput) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	e.logger.Info("Error output: sending stop activity signal")
	e.stopActivityChannel <- true
	return nil
}

func (e *errorOutput) Close(ctx context.Context) error {
	return nil
}
