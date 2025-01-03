package connectiondata

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoDbConnectionDataService struct {
	logger      *slog.Logger
	connection  *mgmtv1alpha1.Connection
	connconfig  *mgmtv1alpha1.MongoConnectionConfig
	mongoclient mongoconnect.Interface
}

func NewMongoDbConnectionDataService(
	logger *slog.Logger,
	connection *mgmtv1alpha1.Connection,
	mongoclient mongoconnect.Interface,
) *MongoDbConnectionDataService {
	return &MongoDbConnectionDataService{
		logger:      logger,
		connection:  connection,
		connconfig:  connection.GetConnectionConfig().GetMongoConfig(),
		mongoclient: mongoclient,
	}
}

func (s *MongoDbConnectionDataService) StreamData(
	ctx context.Context,
	stream *connect.ServerStream[mgmtv1alpha1.GetConnectionDataStreamResponse],
	config *mgmtv1alpha1.ConnectionStreamConfig,
	schema, table string,
) error {
	return errors.ErrUnsupported
}

func (s *MongoDbConnectionDataService) GetSchema(
	ctx context.Context,
	config *mgmtv1alpha1.ConnectionSchemaConfig,
) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	db, err := s.mongoclient.NewFromConnectionConfig(s.connection.GetConnectionConfig(), s.logger)
	if err != nil {
		return nil, err
	}
	mongoconn, err := db.Open(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close(ctx)
	dbnames, err := mongoconn.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	schemas := []*mgmtv1alpha1.DatabaseColumn{}
	for _, dbname := range dbnames {
		collectionNames, err := mongoconn.Database(dbname).ListCollectionNames(ctx, bson.D{})
		if err != nil {
			return nil, err
		}
		for _, collectionName := range collectionNames {
			schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
				Schema: dbname,
				Table:  collectionName,
			})
		}
	}
	return schemas, nil
}
