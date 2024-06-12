package sqlconnect

import (
	context "context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	pgxslog "github.com/nucleuscloud/neosync/backend/internal/pgx-slog"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
)

// interface used by SqlConnector to abstract away the opening and closing of a Pgxpool that includes tunneling
type PgPoolContainer interface {
	Open(context.Context) (pg_queries.DBTX, error)
	Close()
}

type PgPool struct {
	pool *pgxpool.Pool

	details *ConnectionDetails

	// instance of the created tunnel
	tunnel *sshtunnel.Sshtunnel

	dsn string

	logger *slog.Logger
}

func newPgPool(details *ConnectionDetails, logger *slog.Logger) *PgPool {
	return &PgPool{
		details: details,
		logger:  logger,
	}
}

func (s *PgPool) GetDsn() string {
	return s.dsn
}

func (s *PgPool) Open(ctx context.Context) (pg_queries.DBTX, error) {
	if s.details.Tunnel != nil {
		ready, err := s.details.Tunnel.Start(s.logger)
		if err != nil {
			return nil, err
		}
		<-ready

		_, localport := s.details.Tunnel.GetLocalHostPort()
		newPort := int32(localport)
		s.details.GeneralDbConnectConfig.Port = newPort
		dsn := s.details.GeneralDbConnectConfig.String()

		config, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			return nil, fmt.Errorf("unable to parse dsn into pg config: %w", err)
		}
		config.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   pgxslog.NewLogger(s.logger),
			LogLevel: tracelog.LogLevelDebug,
		}

		// set max number of connections.
		if s.details.MaxConnectionLimit != nil {
			config.MaxConns = *s.details.MaxConnectionLimit
		}

		db, err := pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			s.details.Tunnel.Close()
			return nil, err
		}
		s.dsn = dsn
		s.pool = db
		s.tunnel = s.details.Tunnel
		return db, nil
	}

	dsn := s.details.GeneralDbConnectConfig.String()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	config.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   pgxslog.NewLogger(s.logger),
		LogLevel: tracelog.LogLevelDebug,
	}
	// set max number of connections.
	if s.details.MaxConnectionLimit != nil {
		config.MaxConns = *s.details.MaxConnectionLimit
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	s.pool = db
	s.dsn = dsn
	return db, nil
}

func (s *PgPool) Close() {
	if s.pool == nil {
		return
	}
	s.dsn = ""
	db := s.pool
	s.pool = nil
	db.Close()
	if s.tunnel != nil {
		tunnel := s.tunnel
		s.tunnel = nil
		tunnel.Close()
	}
}
