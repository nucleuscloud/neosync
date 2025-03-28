package sqlprovider

import (
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqldbtx"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
)

type Provider struct {
	connector sqlconnect.SqlConnector
}

func NewProvider(
	sqlconnector sqlconnect.SqlConnector,
) *Provider {
	return &Provider{connector: sqlconnector}
}

var _ connectionmanager.ConnectionProvider[neosync_benthos_sql.SqlDbtx] = &Provider{}

type sqlDbtxWrapper struct {
	sqldbtx.DBTX
	close func() error
}

func (s *sqlDbtxWrapper) Close() error {
	return s.close()
}

const defaultConnectionTimeoutSeconds = uint32(10)

func (p *Provider) GetConnectionClient(
	cc *mgmtv1alpha1.ConnectionConfig,
	logger *slog.Logger,
) (neosync_benthos_sql.SqlDbtx, error) {
	container, err := p.connector.NewDbFromConnectionConfig(
		cc,
		logger,
		sqlconnect.WithConnectionTimeout(defaultConnectionTimeoutSeconds),
	)
	if err != nil {
		return nil, err
	}
	dbtx, err := container.Open()
	if err != nil {
		return nil, err
	}
	return &sqlDbtxWrapper{DBTX: dbtx, close: func() error {
		return container.Close()
	}}, nil
}

func (p *Provider) CloseClientConnection(client neosync_benthos_sql.SqlDbtx) error {
	return client.Close()
}
