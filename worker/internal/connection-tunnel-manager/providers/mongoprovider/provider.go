package mongoprovider

import (
	"context"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/worker/internal/connection-tunnel-manager"
	"go.mongodb.org/mongo-driver/mongo"
)

type Provider struct{}

var _ connectiontunnelmanager.ConnectionProvider[*mongo.Client] = &Provider{}

func (p *Provider) GetConnectionDetails(
	cc *mgmtv1alpha1.ConnectionConfig,
	connectionTimeout *uint32,
	logger *slog.Logger,
) (*connectiontunnelmanager.ConnectionDetails, error) {
	return &connectiontunnelmanager.ConnectionDetails{
		GeneralDbConnectConfig: sqlconnect.GeneralDbConnectConfig{},
		Tunnel:                 nil,
	}, nil
}

func (p *Provider) GetConnectionClient(driver, connectionString string, opts any) (*mongo.Client, error) {
	// todo
	client, err := mongo.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	return client, nil
}
