package nucleusdb

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
)

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row

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
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			slog.Warn(err.Error())
		}
	}()

	if err = fn(d.Q.WithTx(tx)); err != nil {
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
