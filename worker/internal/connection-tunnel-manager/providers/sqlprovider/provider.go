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

var _ connectiontunnelmanager.ConnectionProvider[neosync_benthos_sql.SqlDbtx] = &Provider{}

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

func (p *Provider) GetConnectionClient(driver, connectionString string, opts any) (neosync_benthos_sql.SqlDbtx, error) {
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		db.SetMaxOpenConns(1) // todo
	}
	return db, nil
}
