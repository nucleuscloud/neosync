package mongoconnect

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/clienttls"
	"github.com/nucleuscloud/neosync/backend/pkg/sshtunnel"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"golang.org/x/crypto/ssh"
)

const (
	localhost  = "localhost"
	randomPort = 0
)

type Interface interface {
	NewFromConnectionConfig(cc *mgmtv1alpha1.ConnectionConfig, logger *slog.Logger) (DbContainer, error)
}

type DbContainer interface {
	Open(ctx context.Context) (*mongo.Client, error)
	Close(context.Context) error
}

type WrappedMongoClient struct {
	client   *mongo.Client
	clientMu sync.Mutex

	details *ConnectionDetails

	logger *slog.Logger
}

var _ DbContainer = &WrappedMongoClient{}

func newWrappedMongoClient(details *ConnectionDetails, logger *slog.Logger) *WrappedMongoClient {
	return &WrappedMongoClient{details: details, logger: logger}
}

func (w *WrappedMongoClient) Open(ctx context.Context) (*mongo.Client, error) {
	w.clientMu.Lock()
	defer w.clientMu.Unlock()
	if w.client != nil {
		return w.client, nil
	}

	if w.details.Tunnel != nil {
		ready, err := w.details.Tunnel.Start(w.logger)
		if err != nil {
			return nil, err
		}
		<-ready
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(w.details.String()).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	w.client = client
	return client, nil
}

func (w *WrappedMongoClient) Close(ctx context.Context) error {
	w.clientMu.Lock()
	defer w.clientMu.Unlock()
	if w.client == nil {
		return nil
	}
	client := w.client
	w.client = nil
	err := client.Disconnect(ctx)
	if w.details.Tunnel != nil && w.details.Tunnel.IsOpen() {
		w.details.Tunnel.Close()
	}
	return err
}

var _ Interface = &Connector{}

type Connector struct{}

func NewConnector() *Connector {
	return &Connector{}
}

func (c *Connector) NewFromConnectionConfig(
	cc *mgmtv1alpha1.ConnectionConfig,
	logger *slog.Logger,
) (DbContainer, error) {
	if cc == nil {
		return nil, errors.New("cc was nil, expected *mgmtv1alpha1.ConnectionConfig")
	}

	details, err := GetConnectionDetails(cc, clienttls.UpsertCLientTlsFiles, logger)
	if err != nil {
		return nil, err
	}
	wrappedclient := newWrappedMongoClient(details, logger)
	return wrappedclient, nil
}

type ConnectionDetails struct {
	Tunnel  *sshtunnel.Sshtunnel
	Details *connstring.ConnString
}

func (c *ConnectionDetails) GetTunnel() *sshtunnel.Sshtunnel {
	return c.Tunnel
}
func (c *ConnectionDetails) String() string {
	if c.Tunnel != nil && c.Tunnel.IsOpen() {
		localhost, port := c.Tunnel.GetLocalHostPort()
		newDetails := *c.Details
		newDetails.Hosts = []string{fmt.Sprintf("%s:%d", localhost, port)}
		return newDetails.String()
	}
	return c.Details.String()
}

func GetConnectionDetails(
	cc *mgmtv1alpha1.ConnectionConfig,
	handleClientTlsConfig clienttls.ClientTlsFileHandler,
	logger *slog.Logger,
) (*ConnectionDetails, error) {
	if cc == nil {
		return nil, errors.New("cc was nil, expected *mgmtv1alpha1.ConnectionConfig")
	}
	mongoConfig := cc.GetMongoConfig()
	if mongoConfig == nil {
		return nil, fmt.Errorf("mongo config was nil, expected ConnectionConfig to contain valid MongoConfig")
	}

	if mongoConfig.GetClientTls() != nil {
		_, err := handleClientTlsConfig(mongoConfig.GetClientTls())
		if err != nil {
			return nil, err
		}
	}
	tunnelCfg := mongoConfig.GetTunnel()
	if tunnelCfg == nil {
		connDetails, err := getGeneralDbConnectConfigFromMongo(mongoConfig)
		if err != nil {
			return nil, err
		}
		return &ConnectionDetails{
			Details: connDetails,
		}, nil
	}

	destination, err := getEndpointFromMongoConnectionConfig(mongoConfig)
	if err != nil {
		return nil, err
	}
	authmethod, err := sshtunnel.GetTunnelAuthMethodFromSshConfig(tunnelCfg.GetAuthentication())
	if err != nil {
		return nil, err
	}
	var publickey ssh.PublicKey
	if tunnelCfg.GetKnownHostPublicKey() == "" {
		publickey, err = sshtunnel.ParseSshKey(tunnelCfg.GetKnownHostPublicKey())
		if err != nil {
			return nil, err
		}
	}
	tunnel := sshtunnel.New(
		sshtunnel.NewEndpointWithUser(tunnelCfg.GetHost(), int(tunnelCfg.GetPort()), tunnelCfg.GetUser()),
		authmethod,
		destination,
		sshtunnel.NewEndpoint(localhost, randomPort),
		1,
		publickey,
	)
	connDetails, err := getGeneralDbConnectConfigFromMongo(mongoConfig)
	if err != nil {
		return nil, err
	}

	return &ConnectionDetails{
		Tunnel:  tunnel,
		Details: connDetails,
	}, nil
}

func getGeneralDbConnectConfigFromMongo(config *mgmtv1alpha1.MongoConnectionConfig) (*connstring.ConnString, error) {
	dburl := config.GetUrl()
	if dburl == "" {
		return nil, fmt.Errorf("must provide valid mongoconfig url")
	}
	return connstring.ParseAndValidate(dburl)
}

func getEndpointFromMongoConnectionConfig(config *mgmtv1alpha1.MongoConnectionConfig) (*sshtunnel.Endpoint, error) {
	details, err := getGeneralDbConnectConfigFromMongo(config)
	if err != nil {
		return nil, err
	}
	return sshtunnel.NewEndpointWithUser(details.Hosts[0], -1, details.Username), nil
}
