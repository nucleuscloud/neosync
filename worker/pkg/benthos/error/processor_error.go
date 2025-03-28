package neosync_benthos_error

import (
	"context"
	"fmt"

	"github.com/redpanda-data/benthos/v4/public/service"
)

func errorProcessorSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Summary(`Sends stop workflow signal`).
		Field(service.NewStringField("error_msg"))
}

// Registers an processor on a benthos environment called error
func RegisterErrorProcessor(env *service.Environment, stopActivityChannel chan<- error) error {
	return env.RegisterBatchProcessor(
		"error", errorProcessorSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			out, err := newErrorProcessor(conf, mgr, stopActivityChannel)
			if err != nil {
				return nil, err
			}
			return out, nil
		})
}

type errorProcessor struct {
	logger              *service.Logger
	stopActivityChannel chan<- error
	errorMsg            *service.InterpolatedString
}

func newErrorProcessor(
	conf *service.ParsedConfig,
	mgr *service.Resources,
	channel chan<- error,
) (*errorProcessor, error) {
	errMsg, err := conf.FieldInterpolatedString("error_msg")
	if err != nil {
		return nil, err
	}
	return &errorProcessor{
		logger:              mgr.Logger(),
		stopActivityChannel: channel,
		errorMsg:            errMsg,
	}, nil
}

func (r *errorProcessor) ProcessBatch(
	_ context.Context,
	batch service.MessageBatch,
) ([]service.MessageBatch, error) {
	for i := range batch {
		errMsg, err := batch.TryInterpolatedString(i, r.errorMsg)
		if err != nil {
			return nil, fmt.Errorf("error message interpolation error: %w", err)
		}
		// kill activity
		r.logger.Error(
			fmt.Sprintf("Benthos Error processor - sending stop activity signal: %s ", errMsg),
		)
		r.stopActivityChannel <- fmt.Errorf("%s", errMsg)
	}
	return []service.MessageBatch{}, nil
}

func (r *errorProcessor) Close(ctx context.Context) error {
	return nil
}
