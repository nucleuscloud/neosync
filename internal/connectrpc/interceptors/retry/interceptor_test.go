package retry_interceptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/cenkalti/backoff/v5"
	"github.com/stretchr/testify/assert"
)

type mockUnaryFunc struct {
	callCount int
	err       error
}

func (m *mockUnaryFunc) Call(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
	m.callCount++
	return nil, m.err
}

func TestInterceptor_WrapUnary(t *testing.T) {
	t.Run("should retry on unavailable error", func(t *testing.T) {
		mock := &mockUnaryFunc{
			err: connect.NewError(connect.CodeUnavailable, errors.New("service unavailable")),
		}
		interceptor := New(WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithMaxTries(2),
				backoff.WithMaxElapsedTime(30 * time.Second),
			}
		}))

		wrapped := interceptor.WrapUnary(mock.Call)
		_, err := wrapped(context.Background(), &connect.Request[any]{})

		assert.Equal(t, 2, mock.callCount) // Initial + 1 retries
		assert.Error(t, err)
		connectErr := err.(*connect.Error)
		assert.Equal(t, connect.CodeUnavailable, connectErr.Code())
	})

	t.Run("should not retry on non-retryable error", func(t *testing.T) {
		mock := &mockUnaryFunc{
			err: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument")),
		}
		interceptor := New(WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithMaxTries(2),
				backoff.WithMaxElapsedTime(10 * time.Millisecond),
			}
		}))

		wrapped := interceptor.WrapUnary(mock.Call)
		_, err := wrapped(context.Background(), &connect.Request[any]{})
		fmt.Println("err", err)
		assert.Equal(t, 1, mock.callCount)
		assert.Error(t, err)
		connectErr := err.(*connect.Error)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	})

	t.Run("should succeed after retries", func(t *testing.T) {
		mock := &mockUnaryFunc{
			err: nil,
		}
		interceptor := New(WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithMaxTries(2),
				backoff.WithMaxElapsedTime(10 * time.Millisecond),
			}
		}))

		wrapped := interceptor.WrapUnary(mock.Call)
		_, err := wrapped(context.Background(), &connect.Request[any]{})

		assert.Equal(t, 1, mock.callCount)
		assert.NoError(t, err)
	})
}

type mockStreamingClientConn struct {
	sendCallCount    int
	receiveCallCount int
	sendErr          error
	receiveErr       error
}

func (m *mockStreamingClientConn) Spec() connect.Spec           { return connect.Spec{} }
func (m *mockStreamingClientConn) Peer() connect.Peer           { return connect.Peer{} }
func (m *mockStreamingClientConn) RequestHeader() http.Header   { return http.Header{} }
func (m *mockStreamingClientConn) ResponseHeader() http.Header  { return http.Header{} }
func (m *mockStreamingClientConn) ResponseTrailer() http.Header { return http.Header{} }
func (m *mockStreamingClientConn) CloseRequest() error          { return nil }
func (m *mockStreamingClientConn) CloseResponse() error         { return nil }

func (m *mockStreamingClientConn) Send(msg any) error {
	m.sendCallCount++
	return m.sendErr
}

func (m *mockStreamingClientConn) Receive(msg any) error {
	m.receiveCallCount++
	return m.receiveErr
}

func TestInterceptor_WrapStreamingClient(t *testing.T) {
	t.Run("should retry stream operations on unavailable error", func(t *testing.T) {
		mock := &mockStreamingClientConn{
			sendErr:    connect.NewError(connect.CodeUnavailable, errors.New("service unavailable")),
			receiveErr: connect.NewError(connect.CodeUnavailable, errors.New("service unavailable")),
		}

		interceptor := New(WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithMaxTries(2),
				backoff.WithMaxElapsedTime(30 * time.Second),
			}
		}))
		wrapped := interceptor.WrapStreamingClient(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
			return mock
		})

		conn := wrapped(context.Background(), connect.Spec{})
		conn.Send(nil)
		conn.Receive(nil)

		assert.Equal(t, 2, mock.sendCallCount)    // Initial + 2 retries
		assert.Equal(t, 2, mock.receiveCallCount) // Initial + 2 retries
	})

	t.Run("should not retry on non-retryable error", func(t *testing.T) {
		mock := &mockStreamingClientConn{
			sendErr:    connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument")),
			receiveErr: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument")),
		}

		interceptor := New(WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithMaxTries(2),
				backoff.WithMaxElapsedTime(30 * time.Second),
			}
		}))
		wrapped := interceptor.WrapStreamingClient(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
			return mock
		})

		conn := wrapped(context.Background(), connect.Spec{})
		conn.Send(nil)
		conn.Receive(nil)

		assert.Equal(t, 1, mock.sendCallCount)    // Only initial call
		assert.Equal(t, 1, mock.receiveCallCount) // Only initial call
	})
	t.Run("should not retry on success", func(t *testing.T) {
		mock := &mockStreamingClientConn{
			sendErr:    nil,
			receiveErr: nil,
		}

		interceptor := New(WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithMaxTries(2),
				backoff.WithMaxElapsedTime(30 * time.Second),
			}
		}))
		wrapped := interceptor.WrapStreamingClient(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
			return mock
		})

		conn := wrapped(context.Background(), connect.Spec{})
		conn.Send(nil)
		conn.Receive(nil)

		assert.Equal(t, 1, mock.sendCallCount)    // Only initial call
		assert.Equal(t, 1, mock.receiveCallCount) // Only initial call
	})
}

func TestInterceptor_WrapStreamingHandler(t *testing.T) {
	t.Run("should retry on retryable error", func(t *testing.T) {
		callCount := 0
		handlerErr := connect.NewError(connect.CodeUnavailable, errors.New("unavailable"))

		interceptor := New(WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithMaxTries(3),
				backoff.WithMaxElapsedTime(30 * time.Second),
			}
		}))

		handler := interceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			callCount++
			if callCount < 3 {
				return handlerErr
			}
			return nil
		})

		err := handler(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, 3, callCount) // Initial + 2 retries
	})

	t.Run("should not retry on non-retryable error", func(t *testing.T) {
		callCount := 0
		handlerErr := connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))

		interceptor := New(WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithMaxTries(3),
				backoff.WithMaxElapsedTime(30 * time.Second),
			}
		}))

		handler := interceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			callCount++
			return handlerErr
		})

		err := handler(context.Background(), nil)
		assert.Error(t, err)
		assert.Equal(t, handlerErr, err)
		assert.Equal(t, 1, callCount) // Only initial call
	})

	t.Run("should not retry on success", func(t *testing.T) {
		callCount := 0

		interceptor := New(WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithMaxTries(3),
				backoff.WithMaxElapsedTime(30 * time.Second),
			}
		}))

		handler := interceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			callCount++
			return nil
		})

		err := handler(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount) // Only initial call
	})
}
