package pool_sql_provider

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
)

// wrapper used for benthos sql-based connections to retrieve the connection they need
type Provider struct {
	connmanager   connectionmanager.Interface[neosync_benthos_sql.SqlDbtx]
	getConnection func(connectionId string) (connectionmanager.ConnectionInput, error)
	logger        *slog.Logger
	session       connectionmanager.SessionInterface
}

var _ neosync_benthos_sql.ConnectionProvider = (*Provider)(nil)

func NewConnectionProvider(
	connmanager connectionmanager.Interface[neosync_benthos_sql.SqlDbtx],
	getConnection func(connectionId string) (connectionmanager.ConnectionInput, error),
	session connectionmanager.SessionInterface,
	logger *slog.Logger,
) *Provider {
	return &Provider{
		connmanager:   connmanager,
		getConnection: getConnection,
		session:       session,
		logger:        logger,
	}
}

func (p *Provider) GetDb(
	ctx context.Context,
	connectionId string,
) (neosync_benthos_sql.SqlDbtx, error) {
	conn, err := p.getConnection(connectionId)
	if err != nil {
		return nil, err
	}
	return p.connmanager.GetConnection(p.session, conn, p.logger)
}

func (p *Provider) GetDriver(connectionId string) (string, error) {
	conn, err := p.getConnection(connectionId)
	if err != nil {
		return "", err
	}
	switch conn.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return sqlmanager_shared.PostgresDriver, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return sqlmanager_shared.MysqlDriver, nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return sqlmanager_shared.MssqlDriver, nil
	default:
		return "", fmt.Errorf("unsupported connection config when determining driver: %T", conn.GetConnectionConfig().GetConfig())
	}
}
