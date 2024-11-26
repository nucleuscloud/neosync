package providers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
)

type Provider struct {
	mp connectionmanager.ConnectionProvider[neosync_benthos_mongodb.MongoClient]
	sp connectionmanager.ConnectionProvider[neosync_benthos_sql.SqlDbtx]
}

var _ connectionmanager.ConnectionProvider[any] = &Provider{}

func NewProvider(
	mp connectionmanager.ConnectionProvider[neosync_benthos_mongodb.MongoClient],
	sp connectionmanager.ConnectionProvider[neosync_benthos_sql.SqlDbtx],
) *Provider {
	return &Provider{
		mp: mp,
		sp: sp,
	}
}

func (p *Provider) GetConnectionClient(cc *mgmtv1alpha1.ConnectionConfig) (any, error) {
	if cc == nil {
		cc = &mgmtv1alpha1.ConnectionConfig{}
	}
	switch cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return p.mp.GetConnectionClient(cc)
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return p.sp.GetConnectionClient(cc)
	default:
		return nil, fmt.Errorf("unsupported connection config: %T", cc.GetConfig())
	}
}

func (p *Provider) CloseClientConnection(client any) error {
	switch typedclient := client.(type) {
	case neosync_benthos_sql.SqlDbtx:
		return p.sp.CloseClientConnection(typedclient)
	case neosync_benthos_mongodb.MongoClient:
		return p.mp.CloseClientConnection(typedclient)
	default:
		return fmt.Errorf("unsupported client, unable to close connection: %T", client)
	}
}
