package sqlprovider

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

var _ connectiontunnelmanager.ConnectionProvider[neosync_benthos_sql.SqlDbtx] = &Provider{}

func (p *Provider) GetConnectionClient(c *mgmtv1alpha1.ConnectionConfig) (neosync_benthos_sql.SqlDbtx, error) {
	// todo: this needs to now open the tunnel
	// db, err := sql.Open(driver, connectionString)
	// if err != nil {
	// 	return nil, err
	// }
	// if opts != nil && opts.MaxConnectionLimit != nil {
	// 	db.SetMaxOpenConns(int(*opts.MaxConnectionLimit))
	// }
	// return db, nil
	// todo
	return nil, nil
}

func (p *Provider) CloseClientConnection(client neosync_benthos_sql.SqlDbtx) error {
	return client.Close()
}

// func getMaxConnectionLimitFromConnection(cc *mgmtv1alpha1.ConnectionConfig) *int32 {
// 	if cc == nil {
// 		return nil
// 	}
// 	switch config := cc.GetConfig().(type) {
// 	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
// 		if config.MysqlConfig != nil && config.MysqlConfig.ConnectionOptions != nil {
// 			return config.MysqlConfig.ConnectionOptions.MaxConnectionLimit
// 		}
// 		return nil
// 	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
// 		if config.PgConfig != nil && config.PgConfig.ConnectionOptions != nil {
// 			return config.PgConfig.ConnectionOptions.MaxConnectionLimit
// 		}
// 		return nil
// 	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
// 		if config.MssqlConfig != nil && config.MssqlConfig.GetConnectionOptions() != nil {
// 			return config.MssqlConfig.GetConnectionOptions().MaxConnectionLimit
// 		}
// 		return nil
// 	}
// 	return nil
// }
