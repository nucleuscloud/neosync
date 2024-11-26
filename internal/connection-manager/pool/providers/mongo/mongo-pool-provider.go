package pool_mongo_provider

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
)

type Getter func(url string) (neosync_benthos_mongodb.MongoClient, error)

// wrapper used for benthos mongo-based connections to retrieve the connection they need
type Provider struct {
	getter Getter
}

var _ neosync_benthos_mongodb.MongoPoolProvider = (*Provider)(nil)

func NewProvider(getter Getter) *Provider {
	return &Provider{getter: getter}
}

func (p *Provider) GetClient(url string) (neosync_benthos_mongodb.MongoClient, error) {
	return p.getter(url)
}

// Returns a function that converts a raw DSN directly to the relevant pooled sql client.
// Allows sharing connections across activities for effective pooling and SSH tunnel management.
func GetMongoPoolProviderGetter(
	tunnelmanager connectionmanager.Interface[any],
	dsnToConnectionIdMap *sync.Map,
	connectionMap map[string]*mgmtv1alpha1.Connection,
	session string,
	slogger *slog.Logger,
) Getter {
	return func(url string) (neosync_benthos_mongodb.MongoClient, error) {
		connid, ok := dsnToConnectionIdMap.Load(url)
		if !ok {
			return nil, errors.New("unable to find connection id by dsn when getting mongo pool")
		}
		connectionId, ok := connid.(string)
		if !ok {
			return nil, fmt.Errorf("unable to convert connection id to string. Type was %T", connectionId)
		}
		connection, ok := connectionMap[connectionId]
		if !ok {
			return nil, errors.New("unable to find connection by connection id when getting db pool")
		}
		connclient, err := tunnelmanager.GetConnection(session, connection, slogger)
		if err != nil {
			return nil, err
		}
		// tunnel manager is generic and can return all different kinda of database clients.
		// Due to this, we have to make sure it is of the correct type as we expect this to be SQL connections
		dbclient, ok := connclient.(neosync_benthos_mongodb.MongoClient)
		if !ok {
			return nil, fmt.Errorf("unable to convert connection client to neosync_benthos_mongodb.MongoClient. Type was %T", connclient)
		}
		return dbclient, nil
	}
}
