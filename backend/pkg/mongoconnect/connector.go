package mongoconnect

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
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

	details *connstring.ConnString
	// tunnel  *sshtunnel.Sshtunnel

	// logger *slog.Logger
}

var _ DbContainer = &WrappedMongoClient{}

func newWrappedMongoClient(details *connstring.ConnString) *WrappedMongoClient {
	return &WrappedMongoClient{details: details}
}

func (w *WrappedMongoClient) Open(ctx context.Context) (*mongo.Client, error) {
	w.clientMu.Lock()
	defer w.clientMu.Unlock()
	if w.client != nil {
		return w.client, nil
	}
	// todo: tunneling
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
	return client.Disconnect(ctx)
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

	details, err := GetConnectionDetails(cc, clienttls.UpsertClientTlsFileSingleClient, logger)
	if err != nil {
		return nil, err
	}
	wrappedclient := newWrappedMongoClient(details.Details)
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
	// todo: add tunnel support
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

	var destination *sshtunnel.Endpoint // todo
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
	_ = connDetails

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

	if config.GetClientTls() != nil {
		parsedurl, err := url.Parse(dburl)
		if err != nil {
			return nil, err
		}

		filenames := clienttls.GetClientTlsFileNamesSingleClient(config.GetClientTls())
		query := parsedurl.Query()
		query.Set("tls", "true")
		if filenames.RootCert != nil {
			query.Set("tlsCAFile", *filenames.RootCert)
		}
		if filenames.ClientKey != nil && filenames.ClientCert != nil {
			query.Set("tlsCertificateKeyFile", *filenames.ClientKey)
		}
		parsedurl.RawQuery = query.Encode()
		dburl = parsedurl.String()
	}
	return connstring.ParseAndValidate(dburl)
}
