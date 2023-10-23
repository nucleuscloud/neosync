package nucleusdb

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
)

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row

	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)

	Ping(ctx context.Context) error

	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type NucleusDb struct {
	db DBTX
	Q  *db_queries.Queries
}

type ConnectConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Pass     string
	SslMode  *string
}

func New(db DBTX) *NucleusDb {
	return &NucleusDb{
		db: db,
		Q:  db_queries.New(db),
	}
}

func NewFromConfig(config *ConnectConfig) (*NucleusDb, error) {
	pool, err := pgxpool.New(context.Background(), GetDbUrl(config))
	if err != nil {
		return nil, err
	}
	return New(pool), nil
}

func (d *NucleusDb) WithTx(
	ctx context.Context,
	opts *pgx.TxOptions,
	fn func(q *db_queries.Queries) error,
) error {
	tx, err := d.getTx(ctx, opts)
	if err != nil {
		return err
	}
	defer HandlePgxRollback(ctx, tx, slog.Default())

	if err := fn(d.Q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (d *NucleusDb) getTx(
	ctx context.Context,
	opts *pgx.TxOptions,
) (pgx.Tx, error) {
	if opts == nil {
		return d.db.Begin(ctx)
	}
	return d.db.BeginTx(ctx, *opts)
}

type PgxRollbackInterface interface {
	Rollback(context.Context) error
}

// Only logs if error is not ErrTxClosed
func HandlePgxRollback(ctx context.Context, tx PgxRollbackInterface, logger *slog.Logger) {
	if err := tx.Rollback(ctx); err != nil && !isTxDone(err) {
		logger.ErrorContext(ctx, err.Error())
	}
}

type SqlRollbackInterface interface {
	Rollback() error
}

func HandleSqlRollback(
	tx SqlRollbackInterface,
	logger *slog.Logger,
) {
	if err := tx.Rollback(); err != nil && !isTxDone(err) {
		logger.Error(err.Error())
	}
}

func isTxDone(err error) bool {
	return errors.Is(err, pgx.ErrTxClosed) || errors.Is(err, sql.ErrTxDone)
}
