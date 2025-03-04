package backoffutil

import (
	"context"

	"github.com/cenkalti/backoff/v5"
)

// Retry is a helper function that retries an operation with the provided retry options
// and a custom retryable error check function.
//
// It takes a context, a function to execute, a function to get retry options,
// and a function to check if an error is retryable.
func Retry[T any](
	ctx context.Context,
	fn func() (T, error),
	getOpts func() []backoff.RetryOption,
	isRetryable func(error) bool,
) (T, error) {
	opts := getOpts()
	return retryUnwrap(backoff.Retry(ctx, retryWrap(fn, isRetryable), opts...))
}

// wraps the input operation to properly handle retryable errors
func retryWrap[T any](fn func() (T, error), isRetryable func(error) bool) func() (T, error) {
	return func() (T, error) {
		res, err := fn()
		if err != nil {
			return res, handleErrorForRetry(err, isRetryable)
		}
		return res, nil
	}
}

// unwraps the result of a final retryable operation and returns the result and the error
func retryUnwrap[T any](res T, err error) (T, error) {
	if err != nil {
		return res, unwrapPermanentError(err)
	}
	return res, nil
}

// If the error is not retryable, it is wrapped in a PermanentError
func handleErrorForRetry(err error, isRetryable func(error) bool) error {
	if isRetryable(err) {
		return err
	}
	return backoff.Permanent(err)
}

// unwrapPermanentError unwraps a PermanentError and returns the underlying error
func unwrapPermanentError(err error) error {
	if err == nil {
		return nil
	}
	permanentErr, ok := err.(*backoff.PermanentError)
	if !ok {
		return err
	}
	return permanentErr.Unwrap()
}
