package sync_activity

import "database/sql"

type conngetter = func(dsn string) (*sql.DB, error)

type poolProvider struct {
	getter conngetter
}

func newPoolProvider(getter conngetter) *poolProvider {
	return &poolProvider{getter: getter}
}

func (p *poolProvider) GetDb(driver, dsn string) (*sql.DB, error) {
	return p.getter(dsn)
}
