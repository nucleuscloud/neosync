package neosync_benthos_mongodb

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	moFieldCollection = "collection"
	moFieldBatching   = "batching"
	moFieldRetries    = "retries"
)

const (
	crboFieldMaxRetries     = "max_retries"
	crboFieldBackOff        = "backoff"
	crboFieldInitInterval   = "initial_interval"
	crboFieldMaxInterval    = "max_interval"
	crboFieldMaxElapsedTime = "max_elapsed_time"
)

func commonRetryBackOffFields(
	defaultMaxRetries int,
	defaultInitInterval string,
	defaultMaxInterval string,
	defaultMaxElapsed string,
) []*service.ConfigField {
	return []*service.ConfigField{
		service.NewIntField(crboFieldMaxRetries).
			Description("The maximum number of retries before giving up on the request. If set to zero there is no discrete limit.").
			Default(defaultMaxRetries).
			Advanced(),
		service.NewObjectField(crboFieldBackOff,
			service.NewDurationField(crboFieldInitInterval).
				Description("The initial period to wait between retry attempts.").
				Default(defaultInitInterval),
			service.NewDurationField(crboFieldMaxInterval).
				Description("The maximum period to wait between retry attempts.").
				Default(defaultMaxInterval),
			service.NewDurationField(crboFieldMaxElapsedTime).
				Description("The maximum period to wait before retry attempts are abandoned. If zero then no limit is used.").
				Default(defaultMaxElapsed),
		).
			Description("Control time intervals between retry attempts.").
			Advanced(),
	}
}

func outputSpec() *service.ConfigSpec {
	spec := service.NewConfigSpec().
		Version("3.43.0").
		Categories("Services").
		Summary("Inserts items into a MongoDB collection.").
		// Description(output.Description(true, true, "")).
		Fields(clientFields()...).
		Fields(
			service.NewStringField(moFieldCollection).
				Description("The name of the target collection."),
			outputOperationDocs(OperationUpdateOne),
			writeConcernDocs(),
		).
		Fields(writeMapsFields()...).
		Fields(
			service.NewOutputMaxInFlightField(),
			service.NewBatchPolicyField(moFieldBatching),
		)
	for _, f := range commonRetryBackOffFields(3, "1s", "5s", "30s") {
		spec = spec.Field(f.Deprecated())
	}
	return spec
}

func RegisterPooledMongoDbOutput(env *service.Environment, clientProvider MongoPoolProvider) error {
	return env.RegisterBatchOutput(
		"pooled_mongodb",
		outputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (out service.BatchOutput, batchPolicy service.BatchPolicy, maxInFlight int, err error) {
			if batchPolicy, err = conf.FieldBatchPolicy(moFieldBatching); err != nil {
				return
			}
			if maxInFlight, err = conf.FieldMaxInFlight(); err != nil {
				return
			}
			if out, err = newOutputWriter(conf, mgr, clientProvider); err != nil {
				return
			}
			return
		},
	)
}

// ------------------------------------------------------------------------------

type outputWriter struct {
	log *service.Logger

	client                       MongoClient
	database                     *mongo.Database
	collection                   *service.InterpolatedString
	writeConcernCollectionOption *options.CollectionOptions
	operation                    Operation
	writeMaps                    writeMaps

	mu sync.Mutex
}

func newOutputWriter(
	conf *service.ParsedConfig,
	res *service.Resources,
	clientProvider MongoPoolProvider,
) (db *outputWriter, err error) {
	db = &outputWriter{
		log: res.Logger(),
	}
	neosyncConnectionid, err := conf.FieldString(commonFieldClientConnectionId)
	if err != nil {
		return nil, err
	}
	dbname, err := conf.FieldString(commonFieldClientDatabase)
	if err != nil {
		return nil, err
	}
	mClient, err := clientProvider.GetClient(context.Background(), neosyncConnectionid)
	if err != nil {
		return nil, err
	}
	db.client = mClient
	db.database = mClient.Database(dbname)

	if db.collection, err = conf.FieldInterpolatedString(moFieldCollection); err != nil {
		return nil, err
	}
	if db.writeConcernCollectionOption, err = writeConcernCollectionOptionFromParsed(conf); err != nil {
		return nil, err
	}
	if db.operation, err = operationFromParsed(conf); err != nil {
		return nil, err
	}
	if db.writeMaps, err = writeMapsFromParsed(conf, db.operation); err != nil {
		return nil, err
	}
	return db, nil
}

// Connect attempts to establish a connection to the target mongo DB.
func (m *outputWriter) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.client.Ping(ctx, nil); err != nil {
		_ = m.client.Disconnect(ctx)
		return fmt.Errorf("ping failed: %v", err)
	}
	return nil
}

func (m *outputWriter) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	m.mu.Lock()
	collection := m.collection
	m.mu.Unlock()

	if collection == nil {
		return service.ErrNotConnected
	}

	writeModelsMap := map[string][]mongo.WriteModel{}

	err := batch.WalkWithBatchedErrors(func(i int, _ *service.Message) error {
		var err error

		collectionStr, err := batch.TryInterpolatedString(i, collection)
		if err != nil {
			return fmt.Errorf("collection interpolation error: %w", err)
		}

		docJSON, filterJSON, hintJSON, err := m.writeMaps.extractFromMessage(m.operation, i, batch)
		if err != nil {
			return err
		}

		var writeModel mongo.WriteModel
		switch m.operation {
		case OperationInsertOne:
			writeModel = &mongo.InsertOneModel{
				Document: docJSON,
			}
		case OperationDeleteOne:
			writeModel = &mongo.DeleteOneModel{
				Filter: filterJSON,
				Hint:   hintJSON,
			}
		case OperationDeleteMany:
			writeModel = &mongo.DeleteManyModel{
				Filter: filterJSON,
				Hint:   hintJSON,
			}
		case OperationReplaceOne:
			writeModel = &mongo.ReplaceOneModel{
				Upsert:      &m.writeMaps.upsert,
				Filter:      filterJSON,
				Replacement: docJSON,
				Hint:        hintJSON,
			}
		case OperationUpdateOne:
			writeModel = &mongo.UpdateOneModel{
				Upsert: &m.writeMaps.upsert,
				Filter: filterJSON,
				Update: docJSON,
				Hint:   hintJSON,
			}
		}

		if writeModel != nil {
			writeModelsMap[collectionStr] = append(writeModelsMap[collectionStr], writeModel)
		}
		return nil
	})

	// Check for fatal errors and exit immediately if we encounter one
	var batchErr *service.BatchError
	if err != nil {
		if !errors.As(err, &batchErr) {
			return err
		}
	}

	// Dispatch any documents which WalkWithBatchedErrors managed to process successfully
	if len(writeModelsMap) > 0 {
		for collectionStr, writeModels := range writeModelsMap {
			// We should have at least one write model in the slice
			collection := m.database.Collection(collectionStr, m.writeConcernCollectionOption)
			if _, err := collection.BulkWrite(ctx, writeModels); err != nil {
				return err
			}
		}
	}

	// Return any errors produced by invalid messages from the batch
	if batchErr != nil {
		return batchErr
	}
	return nil
}

func (m *outputWriter) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var err error
	if m.client != nil {
		err = m.client.Disconnect(ctx)
		m.client = nil
	}
	m.collection = nil
	return err
}
