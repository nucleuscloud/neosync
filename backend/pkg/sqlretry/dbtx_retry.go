package sqlretry

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/nucleuscloud/neosync/backend/pkg/sqldbtx"
	"github.com/nucleuscloud/neosync/internal/backoffutil"
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
	return backoffutil.Retry(ctx, operation, r.config.getRetryOpts, isRetryableError)
}

func (r *RetryDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	operation := func() (*sql.Stmt, error) {
		return r.dbtx.PrepareContext(ctx, query)
	}
	return backoffutil.Retry(ctx, operation, r.config.getRetryOpts, isRetryableError)
}

func (r *RetryDBTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	operation := func() (*sql.Rows, error) {
		return r.dbtx.QueryContext(ctx, query, args...)
	}
	return backoffutil.Retry(ctx, operation, r.config.getRetryOpts, isRetryableError)
}

func (r *RetryDBTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return r.dbtx.QueryRowContext(ctx, query, args...)
}

func (r *RetryDBTX) PingContext(ctx context.Context) error {
	operation := func() (any, error) {
		return nil, r.dbtx.PingContext(ctx)
	}
	_, err := backoffutil.Retry(ctx, operation, r.config.getRetryOpts, isRetryableError)
	return err
}

func (r *RetryDBTX) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	operation := func() (*sql.Tx, error) {
		return r.dbtx.BeginTx(ctx, opts)
	}
	return backoffutil.Retry(ctx, operation, r.config.getRetryOpts, isRetryableError)
}

func (r *RetryDBTX) RetryTx(ctx context.Context, opts *sql.TxOptions, fn func(*sql.Tx) error) error {
	operation := func() (any, error) {
		tx, err := r.dbtx.BeginTx(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}

		// If fn returns error, rollback
		err = fn(tx)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return nil, fmt.Errorf("error: %v, rollback error: %v", err, rollbackErr)
			}
			return nil, err
		}

		// If fn succeeds, commit
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil, nil
	}

	_, err := backoffutil.Retry(ctx, operation, r.config.getRetryOpts, isRetryableError)
	return err
}

const (
	pqSerializationFailure = "40001" // also means recovery conflict
	pqLockNotAvailable     = "55P03"
	pqObjectInUse          = "55006"
	pqTooManyConnections   = "53300"
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

	if isRetryablePostgresError(err) {
		return true
	}

	if isNetworkError(err) {
		return true
	}

	return false
}

var (
	networkErrors = []string{
		"unexpected eof", // Important for cases that don't explicitly return io.ErrUnexpectedEOF
		"connection reset by peer",
		"broken pipe",
		"connection refused",
		"i/o timeout",
		"no connection",
		"connection closed",
	}
)

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, driver.ErrBadConn) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	for _, netErr := range networkErrors {
		if strings.Contains(errMsg, netErr) {
			return true
		}
	}

	return false
}

func isRetryablePostgresError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pqSerializationFailure,
			pqLockNotAvailable,
			pqObjectInUse,
			pqTooManyConnections:
			return true
		}
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case pqSerializationFailure,
			pqLockNotAvailable,
			pqObjectInUse,
			pqTooManyConnections:
			return true
		}
	}

	return false
}
