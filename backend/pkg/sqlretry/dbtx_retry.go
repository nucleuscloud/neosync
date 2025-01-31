package sqlretry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/nucleuscloud/neosync/backend/pkg/sqldbtx"
)

type RetryDBTX struct {
	dbtx sqldbtx.DBTX

	config *config
}

var _ sqldbtx.DBTX = (*RetryDBTX)(nil)

type config struct {
	getRetryOpts func() []backoff.RetryOption
}

type Option func(*config)

func NewDefault(dbtx sqldbtx.DBTX, logger *slog.Logger) *RetryDBTX {
	return New(dbtx, WithRetryOptions(
		func() []backoff.RetryOption {
			backoffStrategy := backoff.NewExponentialBackOff()
			backoffStrategy.InitialInterval = 200 * time.Millisecond
			backoffStrategy.MaxInterval = 30 * time.Second
			backoffStrategy.Multiplier = 2
			backoffStrategy.RandomizationFactor = 0.3
			return []backoff.RetryOption{
				backoff.WithBackOff(backoffStrategy),
				backoff.WithMaxTries(25),
				backoff.WithMaxElapsedTime(5 * time.Minute),
				backoff.WithNotify(func(err error, d time.Duration) {
					logger.Warn(fmt.Sprintf("sql error with retry: %s, retrying in %s", err.Error(), d.String()))
				}),
			}
		},
	))
}

func noRetryOptions() []backoff.RetryOption {
	return []backoff.RetryOption{}
}

func New(dbtx sqldbtx.DBTX, opts ...Option) *RetryDBTX {
	cfg := &config{getRetryOpts: noRetryOptions}
	for _, opt := range opts {
		opt(cfg)
	}
	return &RetryDBTX{dbtx: dbtx, config: cfg}
}

func WithRetryOptions(getRetryOpts func() []backoff.RetryOption) Option {
	return func(cfg *config) {
		cfg.getRetryOpts = getRetryOpts
	}
}

func (r *RetryDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	operation := func() (sql.Result, error) {
		return r.dbtx.ExecContext(ctx, query, args...)
	}
	return retry(ctx, operation, r.config.getRetryOpts)
}

func (r *RetryDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	operation := func() (*sql.Stmt, error) {
		return r.dbtx.PrepareContext(ctx, query)
	}
	return retry(ctx, operation, r.config.getRetryOpts)
}

func (r *RetryDBTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	operation := func() (*sql.Rows, error) {
		return r.dbtx.QueryContext(ctx, query, args...)
	}
	return retry(ctx, operation, r.config.getRetryOpts)
}

func (r *RetryDBTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return r.dbtx.QueryRowContext(ctx, query, args...)
}

func (r *RetryDBTX) PingContext(ctx context.Context) error {
	operation := func() (any, error) {
		return nil, r.dbtx.PingContext(ctx)
	}
	_, err := retry(ctx, operation, r.config.getRetryOpts)
	return err
}

func (r *RetryDBTX) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	operation := func() (*sql.Tx, error) {
		return r.dbtx.BeginTx(ctx, opts)
	}
	return retry(ctx, operation, r.config.getRetryOpts)
}

func retry[T any](ctx context.Context, fn func() (T, error), getOpts func() []backoff.RetryOption) (T, error) {
	opts := getOpts()
	return retryUnwrap(backoff.Retry(ctx, retryWrap(fn), opts...))
}

// wraps the input operation to properly handle retryable errors
func retryWrap[T any](fn func() (T, error)) func() (T, error) {
	return func() (T, error) {
		res, err := fn()
		if err != nil {
			return res, handleErrorForRetry(err)
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

func handleErrorForRetry(err error) error {
	if isRetryableError(err) {
		return err
	}
	return backoff.Permanent(err)
}

const (
	pqSerializationFailure = "40001"
	pqLockNotAvailable     = "55P03"
	pqObjectInUse          = "55006"
)

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if isDeadlockError(err) {
		return true
	}
	if errors.Is(err, mysql.ErrBusyBuffer) {
		return true
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case pqSerializationFailure,
			pqLockNotAvailable,
			pqObjectInUse:
			return true
		}
	}
	return false
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
