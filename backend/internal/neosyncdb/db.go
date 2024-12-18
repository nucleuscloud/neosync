package neosyncdb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	pgxslog "github.com/nucleuscloud/neosync/backend/internal/pgx-slog"
)

type DBTX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Ping(ctx context.Context) error
	BaseDBTX
}

type BaseDBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row

	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type NeosyncDb struct {
	Db DBTX
	Q  db_queries.Querier
}

type ConnectConfig struct {
	Host                  string
	Port                  int
	Database              string
	User                  string
	Pass                  string
	SslMode               *string
	MigrationsTableName   *string
	MigrationsTableQuoted *bool
	Options               *string
}

func New(db DBTX, q db_queries.Querier) *NeosyncDb {
	if q != nil {
		return &NeosyncDb{
			Db: db,
			Q:  q,
		}
	}
	return &NeosyncDb{
		Db: db,
		Q:  db_queries.New(),
	}
}

func NewPool(config *ConnectConfig) (*pgxpool.Pool, error) {
	pgxconfig, err := pgxpool.ParseConfig(GetDbUrl(config))
	if err != nil {
		return nil, fmt.Errorf("unable to parse pgxpool config: %w", err)
	}
	pgxconfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   pgxslog.NewLogger(slog.Default(), pgxslog.GetShouldOmitArgs()),
		LogLevel: pgxslog.GetDatabaseLogLevel(),
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxconfig)
	if err != nil {
		return nil, fmt.Errorf("uanble to initialize pgx pool from configuration: %w", err)
	}
	return pool, nil
}

func NewFromConfig(config *ConnectConfig) (*NeosyncDb, error) {
	pool, err := NewPool(config)
	if err != nil {
		return nil, err
	}
	return New(pool, nil), nil
}

func (d *NeosyncDb) WithTx(
	ctx context.Context,
	opts *pgx.TxOptions,
	fn func(db BaseDBTX) error,
) error {
	tx, err := d.getTx(ctx, opts)
	if err != nil {
		return err
	}
	defer HandlePgxRollback(ctx, tx, slog.Default())

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (d *NeosyncDb) getTx(
	ctx context.Context,
	opts *pgx.TxOptions,
) (pgx.Tx, error) {
	if opts == nil {
		return d.Db.Begin(ctx)
	}
	return d.Db.BeginTx(ctx, *opts)
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
