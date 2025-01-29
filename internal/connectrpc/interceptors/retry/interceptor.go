package retry_interceptor

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/cenkalti/backoff/v5"
)

type Interceptor struct {
	config *config
}

var _ connect.Interceptor = &Interceptor{}

type config struct {
	getRetryOptions func() []backoff.RetryOption
}

type Option func(*config)

func DefaultRetryInterceptor(logger *slog.Logger) *Interceptor {
	return New(
		WithRetryOptions(func() []backoff.RetryOption {
			return []backoff.RetryOption{
				backoff.WithBackOff(backoff.NewExponentialBackOff()),
				backoff.WithMaxTries(10),
				backoff.WithMaxElapsedTime(1 * time.Minute),
				backoff.WithNotify(func(err error, d time.Duration) {
					logger.Warn(fmt.Sprintf("error with retry: %s, retrying in %s", err.Error(), d.String()))
				}),
			}
		}),
	)
}

func New(opts ...Option) *Interceptor {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return &Interceptor{config: cfg}
}

func WithRetryOptions(getRetryOptions func() []backoff.RetryOption) Option {
	return func(cfg *config) {
		cfg.getRetryOptions = getRetryOptions
	}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		operation := func() (connect.AnyResponse, error) {
			response, err := next(ctx, request)
			if err != nil {
				return nil, handleErrorForRetry(err)
			}
			return response, nil
		}

		opts := i.config.getRetryOptions()
		res, err := backoff.Retry(ctx, operation, opts...)
		if err != nil {
			return nil, unwrapPermanentError(err)
		}
		return res, nil
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)

		// Return a wrapped connection that implements retry logic
		return &retryStreamingClientConn{
			conn:   conn,
			ctx:    ctx,
			config: i.config,
			spec:   spec,
			nextFn: next,
		}
	}
}

// retryStreamingClientConn wraps a StreamingClientConn to add retry functionality
type retryStreamingClientConn struct {
	conn   connect.StreamingClientConn
	ctx    context.Context
	config *config
	spec   connect.Spec
	nextFn connect.StreamingClientFunc
}

// Implement the StreamingClientConn interface
func (r *retryStreamingClientConn) Spec() connect.Spec           { return r.conn.Spec() }
func (r *retryStreamingClientConn) Peer() connect.Peer           { return r.conn.Peer() }
func (r *retryStreamingClientConn) RequestHeader() http.Header   { return r.conn.RequestHeader() }
func (r *retryStreamingClientConn) ResponseHeader() http.Header  { return r.conn.ResponseHeader() }
func (r *retryStreamingClientConn) ResponseTrailer() http.Header { return r.conn.ResponseTrailer() }
func (r *retryStreamingClientConn) CloseRequest() error          { return r.conn.CloseRequest() }
func (r *retryStreamingClientConn) CloseResponse() error         { return r.conn.CloseResponse() }

// Send implements retry logic for the Send method
func (r *retryStreamingClientConn) Send(msg any) error {
	operation := func() (any, error) {
		err := r.conn.Send(msg)
		if err != nil {
			return nil, handleErrorForRetry(err)
		}
		return nil, nil
	}

	opts := r.config.getRetryOptions()
	_, err := backoff.Retry(r.ctx, operation, opts...)
	return unwrapPermanentError(err)
}

// Receive implements retry logic for the Receive method
func (r *retryStreamingClientConn) Receive(msg any) error {
	operation := func() (any, error) {
		err := r.conn.Receive(msg)
		if err != nil {
			return nil, handleErrorForRetry(err)
		}
		return nil, nil
	}

	opts := r.config.getRetryOptions()
	_, err := backoff.Retry(r.ctx, operation, opts...)
	return unwrapPermanentError(err)
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		operation := func() (any, error) {
			err := next(ctx, conn)
			if err != nil {
				return nil, handleErrorForRetry(err)
			}
			return nil, nil
		}

		opts := i.config.getRetryOptions()
		_, err := backoff.Retry(ctx, operation, opts...)
		return unwrapPermanentError(err)
	}
}

func handleErrorForRetry(err error) error {
	if isRetryableError(err) {
		return err
	}
	return backoff.Permanent(err)
}

func isRetryableError(err error) bool {
	connectErr, ok := err.(*connect.Error)
	if ok && connectErr.Code() == connect.CodeUnavailable {
		return true
	}
	return false
}

// unwrapPermanentError unwraps a PermanentError and returns the underlying error
// using errors.As would properly find the ConnectError but clients expect the output
// of a Connect RPC to be a ConnectError, not PermanentError, thus we unwrap it.
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
