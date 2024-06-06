package providers

import (
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/internal/benthos/sql"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager"
	"github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager/providers/sqlprovider"
	"go.mongodb.org/mongo-driver/mongo"
)

type Provider struct {
	mp connectiontunnelmanager.ConnectionProvider[*mongo.Client, any]
	sp connectiontunnelmanager.ConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig]
}

var _ connectiontunnelmanager.ConnectionProvider[any, any] = &Provider{}

func NewProvider(
	mp connectiontunnelmanager.ConnectionProvider[*mongo.Client, any],
	sp connectiontunnelmanager.ConnectionProvider[neosync_benthos_sql.SqlDbtx, *sqlprovider.ConnectionClientConfig],
) *Provider {
	return &Provider{
		mp: mp,
		sp: sp,
	}
}

func (p *Provider) GetConnectionDetails(
	cc *mgmtv1alpha1.ConnectionConfig,
	connectionTimeout *uint32,
	logger *slog.Logger,
) (*connectiontunnelmanager.ConnectionDetails, error) {
	switch cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return p.mp.GetConnectionDetails(cc, connectionTimeout, logger)
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return p.sp.GetConnectionDetails(cc, connectionTimeout, logger)
	default:
		return nil, fmt.Errorf("unsupported connection config: %T", cc.GetConfig())
	}
}

func (p *Provider) GetConnectionClient(driver, connectionString string, opts any) (any, error) {
	switch driver {
	case "mysql", "postgres", "postgresql":
		typedopts, ok := opts.(*sqlprovider.ConnectionClientConfig)
		if !ok {
			return nil, fmt.Errorf("opts was not *sqlprovider.ConnectionClientConfig, was %T", opts)
		}
		return p.sp.GetConnectionClient(driver, connectionString, typedopts)
	case "mongodb", "mongodb+srv":
		return p.mp.GetConnectionClient(driver, connectionString, opts)
	}
	return nil, fmt.Errorf("unsupported driver: %s", driver)
}

func (p *Provider) CloseClientConnection(client any) error {
	switch typedclient := client.(type) {
	case neosync_benthos_sql.SqlDbtx:
		return p.sp.CloseClientConnection(typedclient)
	case *mongo.Client:
		return p.mp.CloseClientConnection(typedclient)
	default:
		return fmt.Errorf("unsupported client, unable to close connection: %T", client)
	}
}

func (p *Provider) GetConnectionClientConfig(cc *mgmtv1alpha1.ConnectionConfig) (any, error) {
	switch cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return p.mp.GetConnectionClientConfig(cc)
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return p.sp.GetConnectionClientConfig(cc)
	default:
		return nil, fmt.Errorf("unsupported connection config: %T", cc.GetConfig())
	}
}
