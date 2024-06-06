package providers

import (
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager"
	"github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager/providers/sqlprovider"
)

type Provider struct{}

var _ connectiontunnelmanager.ConnectionProvider[any] = &Provider{}

func (p *Provider) GetConnectionDetails(
	cc *mgmtv1alpha1.ConnectionConfig,
	connectionTimeout *uint32,
	logger *slog.Logger,
) (*connectiontunnelmanager.ConnectionDetails, error) {
	switch cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		p := mongoprovider.Provider{}
		return p.GetConnectionDetails(cc, connectionTimeout, logger)
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_PgConfig:
		p := sqlprovider.Provider{}
		return p.GetConnectionDetails(cc, connectionTimeout, logger)
	}
	return nil, fmt.Errorf("unsupported connection config: %T", cc.GetConfig())
}

func (p *Provider) GetConnectionClient(driver, connectionString string, opts any) (any, error) {
	switch driver {
	case "mysql", "postgres", "postgresql":
		p := sqlprovider.Provider{}
		return p.GetConnectionClient(driver, connectionString, opts)
	case "mongodb", "mongodb+srv":
		p := mongoprovider.Provider{}
		return p.GetConnectionClient(driver, connectionString, opts)
	}
	return nil, fmt.Errorf("unsupported driver: %s", driver)
}
