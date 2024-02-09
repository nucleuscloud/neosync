package sqlconnect

import (
	"database/sql"
	slog "log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
)

// interface used by SqlConnector to abstract away the opening and closing of a sqldb that includes tunneling
type SqlDbContainer interface {
	Open() (SqlDBTX, error)
	Close() error
}

type SqlDb struct {
	db *sql.DB

	connectionConfig *mgmtv1alpha1.ConnectionConfig
	tunnel           *sshtunnel.Sshtunnel
	logger           *slog.Logger

	connectionTimeout *uint32

	dsn string
}

func (s *SqlDb) Open() (SqlDBTX, error) {
	details, err := GetConnectionDetails(s.connectionConfig, s.connectionTimeout, s.logger)
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
		db, err := sql.Open(details.GeneralDbConnectConfig.Driver, dsn)
		if err != nil {
			details.Tunnel.Close()
			return nil, err
		}
		s.db = db
		s.dsn = dsn
		s.tunnel = details.Tunnel
		return db, nil
	}
	dsn := details.GeneralDbConnectConfig.String()
	db, err := sql.Open(details.GeneralDbConnectConfig.Driver, dsn)
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
