package sqlretry

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
)

type RetryDBTX struct {
	dbtx mysql_queries.DBTX

	config *config
}

var _ mysql_queries.DBTX = (*RetryDBTX)(nil)

type config struct {
	retryOpts []backoff.RetryOption
}

type Option func(*config)

func New(dbtx mysql_queries.DBTX, opts ...Option) *RetryDBTX {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return &RetryDBTX{dbtx: dbtx, config: cfg}
}

func WithRetryOptions(opts ...backoff.RetryOption) Option {
	return func(cfg *config) {
		cfg.retryOpts = opts
	}
}

func (r *RetryDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	operation := func() (sql.Result, error) {
		return r.dbtx.ExecContext(ctx, query, args...)
	}
	return retry(ctx, operation, r.config.retryOpts...)
}

func (r *RetryDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	operation := func() (*sql.Stmt, error) {
		return r.dbtx.PrepareContext(ctx, query)
	}
	return retry(ctx, operation, r.config.retryOpts...)
}

func (r *RetryDBTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	operation := func() (*sql.Rows, error) {
		return r.dbtx.QueryContext(ctx, query, args...)
	}
	return retry(ctx, operation, r.config.retryOpts...)
}

func (r *RetryDBTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return r.dbtx.QueryRowContext(ctx, query, args...)
}

func retry[T any](ctx context.Context, fn func() (T, error), opts ...backoff.RetryOption) (T, error) {
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
