package sqlconnect

import (
	"database/sql"
	slog "log/slog"

	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
)

// interface used by SqlConnector to abstract away the opening and closing of a sqldb that includes tunneling
type SqlDbContainer interface {
	Open() (SqlDBTX, error)
	Close() error
}

type SqlDb struct {
	db *sql.DB

	details *ConnectionDetails

	tunnel *sshtunnel.Sshtunnel

	dsn string

	logger *slog.Logger
}

func newSqlDb(details *ConnectionDetails, logger *slog.Logger) *SqlDb {
	return &SqlDb{details: details, logger: logger}
}

func (s *SqlDb) Open() (SqlDBTX, error) {
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
		db, err := sql.Open(s.details.GeneralDbConnectConfig.Driver, dsn)
		if err != nil {
			s.details.Tunnel.Close()
			return nil, err
		}
		s.db = db
		s.dsn = dsn
		s.tunnel = s.details.Tunnel
		return db, nil
	}
	dsn := s.details.GeneralDbConnectConfig.String()
	db, err := sql.Open(s.details.GeneralDbConnectConfig.Driver, dsn)
	s.db = db
	if err != nil {
		return nil, err
	}
	s.dsn = dsn
	return db, nil
}

func (s *SqlDb) GetDsn() string {
	return s.dsn
}

func (s *SqlDb) Close() error {
	if s.db == nil {
		return nil
	}
	db := s.db
	s.dsn = ""
	s.db = nil
	err := db.Close()
	if s.tunnel != nil {
		s.tunnel.Close()
		s.tunnel = nil
	}
	return err
}
