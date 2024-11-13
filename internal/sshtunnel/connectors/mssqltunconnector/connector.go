package mssqltunconnector

import (
	"database/sql/driver"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/nucleuscloud/neosync/internal/sshtunnel"
)

type Connector struct {
	driver.Connector
}

var _ driver.Connector = (*Connector)(nil)

func New(dialer sshtunnel.Dialer, dsn string) (*Connector, func(), error) {
	connector, err := mssql.NewConnector(dsn)
	if err != nil {
		return nil, nil, err
	}

	connector.Dialer = mssql.Dialer(dialer)

	return &Connector{Connector: connector}, func() {}, nil
}
