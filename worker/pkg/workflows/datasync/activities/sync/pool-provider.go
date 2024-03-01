package sync_activity

import (
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
)

type conngetter = func(dsn string) (mysql_queries.DBTX, error)

type poolProvider struct {
	getter conngetter
}

func newPoolProvider(getter conngetter) *poolProvider {
	return &poolProvider{getter: getter}
}

func (p *poolProvider) GetDb(driver, dsn string) (mysql_queries.DBTX, error) {
	return p.getter(dsn)
}
