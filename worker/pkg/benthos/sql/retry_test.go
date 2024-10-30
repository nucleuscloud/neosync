package neosync_benthos_sql

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRetryWithConfig(t *testing.T) {
	t.Run("succeeds_first_try", func(t *testing.T) {
		// Given
		calls := 0
		mockLogger := &mockLogger{}

		operation := func(ctx context.Context) error {
			calls++
			return nil
		}

		config := &retryConfig{
			MaxAttempts: 3,
			RetryDelay:  time.Millisecond,
			Logger:      mockLogger,
			ShouldRetry: func(err error) bool {
				return err.Error() == "temporary error"
			},
		}

		// When
		err := retryWithConfig(context.Background(), config, operation)

		// Then
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if calls != 1 {
			t.Errorf("expected 1 call, got %d", calls)
		}
	})

	t.Run("succeeds_after_retry", func(t *testing.T) {
		// Given
		calls := 0
		mockLogger := &mockLogger{}

		operation := func(ctx context.Context) error {
			calls++
			if calls == 1 {
				return fmt.Errorf("temporary error")
			}
			return nil
		}

		config := &retryConfig{
			MaxAttempts: 3,
			RetryDelay:  time.Millisecond,
			Logger:      mockLogger,
			ShouldRetry: func(err error) bool {
				return err.Error() == "temporary error"
			},
		}

		// When
		err := retryWithConfig(context.Background(), config, operation)

		// Then
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if calls != 2 {
			t.Errorf("expected 2 calls, got %d", calls)
		}
	})

	t.Run("fails_non_retryable", func(t *testing.T) {
		// Given
		calls := 0
		mockLogger := &mockLogger{}

		operation := func(ctx context.Context) error {
			calls++
			return fmt.Errorf("permanent error")
		}

		config := &retryConfig{
			MaxAttempts: 3,
			RetryDelay:  time.Millisecond,
			Logger:      mockLogger,
			ShouldRetry: func(err error) bool {
				return err.Error() == "temporary error"
			},
		}

		// When
		err := retryWithConfig(context.Background(), config, operation)

		// Then
		if err == nil || !strings.Contains(err.Error(), "permanent error") {
			t.Errorf("expected permanent error, got %v", err)
		}
		if calls != 1 {
			t.Errorf("expected 1 call, got %d", calls)
		}
	})

	t.Run("exceeds_max_attempts", func(t *testing.T) {
		// Given
		calls := 0
		mockLogger := &mockLogger{}

		operation := func(ctx context.Context) error {
			calls++
			return fmt.Errorf("temporary error")
		}

		config := &retryConfig{
			MaxAttempts: 2,
			RetryDelay:  time.Millisecond,
			Logger:      mockLogger,
			ShouldRetry: func(err error) bool {
				return err.Error() == "temporary error"
			},
		}

		// When
		err := retryWithConfig(context.Background(), config, operation)

		// Then
		if err == nil || !strings.Contains(err.Error(), "max retry attempts reached") {
			t.Errorf("expected max retry attempts error, got %v", err)
		}
		if calls != 2 {
			t.Errorf("expected 2 calls, got %d", calls)
		}
	})

	t.Run("context_canceled_during_operation", func(t *testing.T) {
		// Given
		calls := 0
		mockLogger := &mockLogger{}

		operation := func(ctx context.Context) error {
			calls++
			return context.Canceled
		}

		config := &retryConfig{
			MaxAttempts: 3,
			RetryDelay:  time.Millisecond,
			Logger:      mockLogger,
			ShouldRetry: func(err error) bool {
				return err.Error() == "temporary error"
			},
		}

		// When
		err := retryWithConfig(context.Background(), config, operation)

		// Then
		if err != context.Canceled {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
		if calls != 1 {
			t.Errorf("expected 1 call, got %d", calls)
		}
	})

	t.Run("context_canceled_during_sleep", func(t *testing.T) {
		// Given
		calls := 0
		mockLogger := &mockLogger{}

		operation := func(ctx context.Context) error {
			calls++
			return fmt.Errorf("temporary error")
		}

		config := &retryConfig{
			MaxAttempts: 3,
			RetryDelay:  200 * time.Millisecond, // Long enough delay to ensure we can cancel
			Logger:      mockLogger,
			ShouldRetry: func(err error) bool {
				return err.Error() == "temporary error"
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		// Cancel the context after a short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		defer cancel()

		// When
		err := retryWithConfig(ctx, config, operation)

		// Then
		expectedErr := "encountered error while sleeping during retry delay: context canceled"
		if err == nil || !strings.Contains(err.Error(), expectedErr) {
			t.Errorf("expected error containing %q, got %v", expectedErr, err)
		}
		if calls != 1 {
			t.Errorf("expected 1 call, got %d", calls)
		}
	})
}

// mockLogger for testing
type mockLogger struct {
	messages []string
}

func (m *mockLogger) Warnf(format string, args ...any) {
	m.messages = append(m.messages, fmt.Sprintf(format, args...))
}

func TestSleepContext(t *testing.T) {
	t.Run("returns_immediately_for_zero_duration", func(t *testing.T) {
		// Given
		ctx := context.Background()
		duration := time.Duration(0)

		// When
		start := time.Now()
		err := sleepContext(ctx, duration)
		elapsed := time.Since(start)

		// Then
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if elapsed > time.Millisecond {
			t.Errorf("expected immediate return, but took %v", elapsed)
		}
	})

	t.Run("returns_immediately_for_negative_duration", func(t *testing.T) {
		// Given
		ctx := context.Background()
		duration := -time.Second

		// When
		start := time.Now()
		err := sleepContext(ctx, duration)
		elapsed := time.Since(start)

		// Then
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if elapsed > time.Millisecond {
			t.Errorf("expected immediate return, but took %v", elapsed)
		}
	})

	t.Run("sleeps_for_specified_duration", func(t *testing.T) {
		// Given
		ctx := context.Background()
		duration := 100 * time.Millisecond

		// When
		start := time.Now()
		err := sleepContext(ctx, duration)
		elapsed := time.Since(start)

		// Then
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if elapsed < duration {
			t.Errorf("sleep duration too short: expected >= %v, got %v", duration, elapsed)
		}
		// Allow some buffer for scheduling delays
		if elapsed > duration*2 {
			t.Errorf("sleep duration too long: expected <= %v, got %v", duration*2, elapsed)
		}
	})

	t.Run("returns_error_when_context_already_canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		duration := 100 * time.Millisecond

		// When
		start := time.Now()
		err := sleepContext(ctx, duration)
		elapsed := time.Since(start)

		// Then
		if err != context.Canceled {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
		if elapsed > time.Millisecond {
			t.Errorf("expected immediate return, but took %v", elapsed)
		}
	})

	t.Run("returns_error_when_context_canceled_during_sleep", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		duration := 200 * time.Millisecond

		// Cancel context after a short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		// When
		start := time.Now()
		err := sleepContext(ctx, duration)
		elapsed := time.Since(start)

		// Then
		if err != context.Canceled {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
		// Should have returned early due to cancellation
		if elapsed >= duration {
			t.Errorf("expected early return before %v, took %v", duration, elapsed)
		}
		// Should have waited for the cancellation
		if elapsed < 45*time.Millisecond {
			t.Errorf("returned too early, expected at least 45ms, got %v", elapsed)
		}
	})

	t.Run("handles_context_with_deadline", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		duration := 200 * time.Millisecond

		// When
		start := time.Now()
		err := sleepContext(ctx, duration)
		elapsed := time.Since(start)

		// Then
		if err != context.DeadlineExceeded {
			t.Errorf("expected context.DeadlineExceeded error, got %v", err)
		}
		// Should have returned early due to deadline
		if elapsed >= duration {
			t.Errorf("expected early return before %v, took %v", duration, elapsed)
		}
		// Should have waited until close to the deadline
		if elapsed < 45*time.Millisecond {
			t.Errorf("returned too early, expected at least 45ms, got %v", elapsed)
		}
	})
}
