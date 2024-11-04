package neosync_benthos_sql

import (
	"context"
	"fmt"
	"time"
)

// retryConfig holds the configuration for retry attempts
type retryConfig struct {
	MaxAttempts uint
	RetryDelay  time.Duration
	Logger      retryLogger
	// ShouldRetry determines if a given error should trigger a retry
	ShouldRetry func(error) bool
}

// retryLogger interface to allow for different logging implementations
type retryLogger interface {
	Warnf(format string, args ...any)
}

// retryOperation represents an operation that can be retried
type retryOperation func(ctx context.Context) error

// retryWithConfig executes an operation with retry logic for deadlock errors
func retryWithConfig(ctx context.Context, config *retryConfig, operation retryOperation) error {
	var lastErr error
	for attempt := uint(0); attempt < config.MaxAttempts; attempt++ {
		err := operation(ctx)
		if err == nil {
			return nil
		}

		if !config.ShouldRetry(err) {
			return err
		}

		lastErr = err
		config.Logger.Warnf("attempt failed (%d/%d). Retrying in %v... Error: %v",
			attempt+1, config.MaxAttempts, config.RetryDelay, err)

		if err := sleepContext(ctx, config.RetryDelay); err != nil {
			return fmt.Errorf("encountered error while sleeping during retry delay: %w", err)
		}
	}

	return fmt.Errorf("max retry attempts reached: %w", lastErr)
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
