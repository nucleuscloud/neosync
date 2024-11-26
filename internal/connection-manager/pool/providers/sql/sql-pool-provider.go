package pool_sql_provider

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
)

type Getter = func(dsn string) (neosync_benthos_sql.SqlDbtx, error)

// wrapper used for benthos sql-based connections to retrieve the connection they need
type Provider struct {
	getter Getter
}

var _ neosync_benthos_sql.DbPoolProvider = (*Provider)(nil)

func NewProvider(getter Getter) *Provider {
	return &Provider{getter: getter}
}

func (p *Provider) GetDb(driver, dsn string) (neosync_benthos_sql.SqlDbtx, error) {
	return p.getter(dsn)
}

// Returns a function that converts a raw DSN directly to the relevant pooled sql client.
// Allows sharing connections across activities for effective pooling and SSH tunnel management.
func GetSqlPoolProviderGetter(
	tunnelmanager connectionmanager.Interface[any],
	dsnToConnectionIdMap *sync.Map,
	connectionMap map[string]*mgmtv1alpha1.Connection,
	session string,
	slogger *slog.Logger,
) Getter {
	return func(dsn string) (neosync_benthos_sql.SqlDbtx, error) {
		connid, ok := dsnToConnectionIdMap.Load(dsn)
		if !ok {
			return nil, errors.New("unable to find connection id by dsn when getting db pool")
		}
		connectionId, ok := connid.(string)
		if !ok {
			return nil, fmt.Errorf("unable to convert connection id to string. Type was %T", connectionId)
		}
		connection, ok := connectionMap[connectionId]
		if !ok {
			return nil, errors.New("unable to find connection by connection id when getting db pool")
		}
		connclient, err := tunnelmanager.GetConnection(session, connection, slogger)
		if err != nil {
			return nil, err
		}
		// tunnel manager is generic and can return all different kinda of database clients.
		// Due to this, we have to make sure it is of the correct type as we expect this to be SQL connections
		dbclient, ok := connclient.(neosync_benthos_sql.SqlDbtx)
		if !ok {
			return nil, fmt.Errorf("unable to convert connection client to neosync_benthos_sql.SqlDbtx. Type was %T", connclient)
		}
		return dbclient, nil
	}
}
