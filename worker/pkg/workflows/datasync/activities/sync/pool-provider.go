package sync_activity

import neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"

type conngetter = func(dsn string) (neosync_benthos_sql.SqlDbtx, error)

type poolProvider struct {
	getter conngetter
}

func newPoolProvider(getter conngetter) *poolProvider {
	return &poolProvider{getter: getter}
}

func (p *poolProvider) GetDb(driver, dsn string) (neosync_benthos_sql.SqlDbtx, error) {
	return p.getter(dsn)
}
