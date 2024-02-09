package sqlconnect

import (
	context "context"
	slog "log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
)

// interface used by SqlConnector to abstract away the opening and closing of a Pgxpool that includes tunneling
type PgPoolContainer interface {
	Open(context.Context) (pg_queries.DBTX, error)
	Close()
}

type PgPool struct {
	pool *pgxpool.Pool

	connectionConfig *mgmtv1alpha1.PostgresConnectionConfig
	tunnel           *sshtunnel.Sshtunnel
	logger           *slog.Logger

	connectionTimeout *uint32

	dsn string
}

func (s *PgPool) GetDsn() string {
	return s.dsn
}

func (s *PgPool) Open(ctx context.Context) (pg_queries.DBTX, error) {
	details, err := GetConnectionDetails(&mgmtv1alpha1.ConnectionConfig{
		Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
			PgConfig: s.connectionConfig,
		},
	}, s.connectionTimeout, s.logger)
	if err != nil {
		return nil, err
	}
	if details.Tunnel != nil {
		ready, err := details.Tunnel.Start()
		if err != nil {
			return nil, err
		}
		<-ready
		newPort := int32(details.Tunnel.Local.Port)
		details.GeneralDbConnectConfig.Port = newPort
		dsn := details.GeneralDbConnectConfig.String()
		db, err := pgxpool.New(ctx, dsn)
		if err != nil {
			details.Tunnel.Close()
			return nil, err
		}
		s.dsn = dsn
		s.pool = db
		s.tunnel = details.Tunnel
		return db, nil
	}

	dsn := details.GeneralDbConnectConfig.String()
	db, err := pgxpool.New(ctx, dsn)
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
