package neosync_benthos_error

import (
	"context"
	"errors"
	"fmt"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func errorOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Summary(`Sends stop Activity signal`).
		Field(service.NewStringField("error_msg")).
		Field(service.NewIntField("max_in_flight").Default(64)).
		Field(service.NewBoolField("is_generate_job").Default(false)).
		Field(service.NewBatchPolicyField("batching"))
}

// Registers an output on a benthos environment called error
func RegisterErrorOutput(env *service.Environment, stopActivityChannel chan<- error) error {
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
			out, err := newErrorOutput(conf, mgr, stopActivityChannel)
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			return out, batchPolicy, maxInFlight, nil
		})
}

func newErrorOutput(conf *service.ParsedConfig, mgr *service.Resources, channel chan<- error) (*errorOutput, error) {
	errMsg, err := conf.FieldInterpolatedString("error_msg")
	if err != nil {
		return nil, err
	}
	isGenerateJob, err := conf.FieldBool("is_generate_job")
	if err != nil {
		return nil, err
	}
	return &errorOutput{
		logger:              mgr.Logger(),
		stopActivityChannel: channel,
		errorMsg:            errMsg,
		isGenerateJob:       isGenerateJob,
	}, nil
}

type errorOutput struct {
	logger              *service.Logger
	stopActivityChannel chan<- error
	errorMsg            *service.InterpolatedString
	isGenerateJob       bool
}

func (e *errorOutput) Connect(ctx context.Context) error {
	return nil
}

func (e *errorOutput) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	if len(batch) > 0 {
		errMsg, err := batch.TryInterpolatedString(0, e.errorMsg)
		if err != nil {
			return fmt.Errorf("error message interpolation error: %w", err)
		}
		if !e.isCriticalError(errMsg) {
			// throw error so that benthos retries
			return errors.New(errMsg)
		}
		// kill activity
		e.logger.Error(fmt.Sprintf("Benthos Error output - sending stop activity signal: %s ", errMsg))
		e.stopActivityChannel <- fmt.Errorf("%s", errMsg)
	}
	return nil
}

func (e *errorOutput) Close(ctx context.Context) error {
	return nil
}

func (e *errorOutput) isCriticalError(errMsg string) bool {
	if e.isGenerateJob {
		return neosync_benthos.IsGenerateJobCriticalError(errMsg)
	}
	return neosync_benthos.IsCriticalError(errMsg)
}
