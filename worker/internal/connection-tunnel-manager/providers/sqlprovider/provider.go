package sqlprovider

import (
	"database/sql"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/internal/benthos/sql"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

var _ connectiontunnelmanager.ConnectionProvider[neosync_benthos_sql.SqlDbtx, *ConnectionClientConfig] = &Provider{}

func (p *Provider) GetConnectionDetails(
	cc *mgmtv1alpha1.ConnectionConfig,
	connectionTimeout *uint32,
	logger *slog.Logger,
) (*connectiontunnelmanager.ConnectionDetails, error) {
	details, err := sqlconnect.GetConnectionDetails(cc, connectionTimeout, sqlconnect.UpsertCLientTlsFiles, logger)
	if err != nil {
		return nil, err
	}
	return &connectiontunnelmanager.ConnectionDetails{
		GeneralDbConnectConfig: details.GeneralDbConnectConfig,
		Tunnel:                 details.Tunnel,
	}, nil
}

type ConnectionClientConfig struct {
	MaxConnectionLimit *int32
}

func (p *Provider) GetConnectionClient(driver, connectionString string, opts *ConnectionClientConfig) (neosync_benthos_sql.SqlDbtx, error) {
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}
	if opts != nil && opts.MaxConnectionLimit != nil {
		db.SetMaxOpenConns(int(*opts.MaxConnectionLimit))
	}
	return db, nil
}

func (p *Provider) CloseClientConnection(client neosync_benthos_sql.SqlDbtx) error {
	return client.Close()
}

func (p *Provider) GetConnectionClientConfig(cc *mgmtv1alpha1.ConnectionConfig) (*ConnectionClientConfig, error) {
	return &ConnectionClientConfig{
		MaxConnectionLimit: getMaxConnectionLimitFromConnection(cc),
	}, nil
}

func getMaxConnectionLimitFromConnection(cc *mgmtv1alpha1.ConnectionConfig) *int32 {
	if cc == nil {
		return nil
	}
	switch config := cc.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		if config.MysqlConfig != nil && config.MysqlConfig.ConnectionOptions != nil {
			return config.MysqlConfig.ConnectionOptions.MaxConnectionLimit
		}
		return nil
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if config.PgConfig != nil && config.PgConfig.ConnectionOptions != nil {
			return config.PgConfig.ConnectionOptions.MaxConnectionLimit
		}
		return nil
	}
	return nil
}
