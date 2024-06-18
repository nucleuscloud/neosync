package mongoprovider

import (
	"context"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/clienttls"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

var _ connectiontunnelmanager.ConnectionProvider[neosync_benthos_mongodb.MongoClient, any] = &Provider{}

func (p *Provider) GetConnectionDetails(
	cc *mgmtv1alpha1.ConnectionConfig,
	connectionTimeout *uint32,
	logger *slog.Logger,
) (connectiontunnelmanager.ConnectionDetails, error) {
	return mongoconnect.GetConnectionDetails(cc, clienttls.UpsertCLientTlsFiles, logger)
}

// this is currently untested as it isn't really used anywhere
func (p *Provider) GetConnectionClient(driver, connectionString string, opts any) (neosync_benthos_mongodb.MongoClient, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts2 := options.Client().ApplyURI(connectionString).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.Background(), opts2)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (p *Provider) CloseClientConnection(client neosync_benthos_mongodb.MongoClient) error {
	return client.Disconnect(context.Background())
}

func (p *Provider) GetConnectionClientConfig(cc *mgmtv1alpha1.ConnectionConfig) (any, error) {
	var result any
	return result, nil
}
