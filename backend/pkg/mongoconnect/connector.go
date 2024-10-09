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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
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

	details, err := getConnectionDetails(cc, clienttls.UpsertClientTlsFileSingleClient)
	if err != nil {
		return nil, err
	}
	wrappedclient := newWrappedMongoClient(details.Details)
	return wrappedclient, nil
}

type ConnectionDetails struct {
	Details *connstring.ConnString
}

func (c *ConnectionDetails) String() string {
	return c.Details.String()
}

func getConnectionDetails(
	cc *mgmtv1alpha1.ConnectionConfig,
	handleClientTlsConfig clienttls.ClientTlsFileHandler,
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
	if tunnelCfg != nil {
		return nil, fmt.Errorf("tunneling in mongodb is not currently supported: %w", errors.ErrUnsupported)
	}

	connDetails, err := getGeneralDbConnectConfigFromMongo(mongoConfig)
	if err != nil {
		return nil, err
	}
	return &ConnectionDetails{
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
		if !query.Has("tls") {
			query.Set("tls", "true")
		}
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
