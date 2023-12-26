package sqlconnect

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlConnector interface {
	Open(driverName, connectionStr string) (*sql.DB, error)
	MysqlOpen(connectionStr string) (*sql.DB, error)
	PgPoolOpen(ctx context.Context, connectionStr string) (*pgxpool.Pool, error)
}

type SqlOpenConnector struct{}

func (rc *SqlOpenConnector) Open(driverName, connectionStr string) (*sql.DB, error) {
	return sql.Open(driverName, connectionStr)
}

func (rc *SqlOpenConnector) MysqlOpen(connectionStr string) (*sql.DB, error) {
	return sql.Open("mysql", connectionStr)
}

func (rc *SqlOpenConnector) PgPoolOpen(ctx context.Context, connectionStr string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, connectionStr)
}
