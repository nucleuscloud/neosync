package neosync_benthos_error

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/benthosdev/benthos/v4/public/service"
)

func errorOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Summary(`Sends stop Activity signal`).
		Field(service.NewStringField("error_msg")).
		Field(service.NewIntField("max_in_flight").Default(64)).
		Field(service.NewBatchPolicyField("batching"))
}

// Registers an output on a benthos environment called error
func RegisterErrorOutput(env *service.Environment, stopActivityChannel chan error) error {
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

func newErrorOutput(conf *service.ParsedConfig, mgr *service.Resources, channel chan error) (*errorOutput, error) {
	errMsg, err := conf.FieldInterpolatedString("error_msg")
	if err != nil {
		return nil, err
	}
	return &errorOutput{
		logger:              mgr.Logger(),
		stopActivityChannel: channel,
		errorMsg:            errMsg,
	}, nil
}

type errorOutput struct {
	logger              *service.Logger
	stopActivityChannel chan error
	errorMsg            *service.InterpolatedString
}

func (e *errorOutput) Connect(ctx context.Context) error {
	return nil
}

func (e *errorOutput) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	for i := range batch {
		errMsg, err := batch.TryInterpolatedString(i, e.errorMsg)
		if err != nil {
			return fmt.Errorf("error message interpolation error: %w", err)
		}
		if isMaxConnectionError(errMsg) {
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

// checks if the error message matches a max connections error
func isMaxConnectionError(errMsg string) bool {
	// list of known error messages for when max connections are reached
	maxConnErrors := []string{
		"too many clients already",
		"remaining connection slots are reserved",
		"maximum number of connections reached",
	}

	for _, errStr := range maxConnErrors {
		if containsIgnoreCase(errMsg, errStr) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
