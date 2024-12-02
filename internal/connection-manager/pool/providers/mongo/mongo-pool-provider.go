package pool_mongo_provider

import (
	"context"
	"log/slog"

	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
)

// wrapper used for benthos mongo-based connections to retrieve the connection they need
type Provider struct {
	connmanager   connectionmanager.Interface[neosync_benthos_mongodb.MongoClient]
	getConnection func(connectionId string) (connectionmanager.ConnectionInput, error)
	logger        *slog.Logger
	session       connectionmanager.SessionInterface
}

var _ neosync_benthos_mongodb.MongoPoolProvider = (*Provider)(nil)

func NewProvider(
	connmanager connectionmanager.Interface[neosync_benthos_mongodb.MongoClient],
	getConnection func(connectionId string) (connectionmanager.ConnectionInput, error),
	session connectionmanager.SessionInterface,
	logger *slog.Logger,
) *Provider {
	return &Provider{connmanager: connmanager, getConnection: getConnection, session: session, logger: logger}
}

func (p *Provider) GetClient(ctx context.Context, connectionId string) (neosync_benthos_mongodb.MongoClient, error) {
	conn, err := p.getConnection(connectionId)
	if err != nil {
		return nil, err
	}
	return p.connmanager.GetConnection(p.session, conn, p.logger)
}
