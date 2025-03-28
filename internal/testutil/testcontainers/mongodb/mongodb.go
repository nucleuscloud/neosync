package testcontainers_mongodb

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

type MongoDBTestSyncContainer struct {
	Source *MongoDBTestContainer
	Target *MongoDBTestContainer
}

func NewMongoDBTestSyncContainer(
	ctx context.Context,
	t *testing.T,
) (*MongoDBTestSyncContainer, error) {
	tc := &MongoDBTestSyncContainer{}
	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		d, err := NewMongoDBTestContainer(ctx, t)
		if err != nil {
			return err
		}
		tc.Source = d
		return nil
	})

	errgrp.Go(func() error {
		d, err := NewMongoDBTestContainer(ctx, t)
		if err != nil {
			return err
		}
		tc.Target = d
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return tc, nil
}

func (d *MongoDBTestSyncContainer) TearDown(ctx context.Context) error {
	if d.Source != nil {
		if d.Source.TestContainer != nil {
			err := d.Source.TestContainer.Terminate(ctx)
			if err != nil {
				return err
			}
		}
	}
	if d.Target != nil {
		if d.Target.TestContainer != nil {
			err := d.Target.TestContainer.Terminate(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type MongoDBTestContainer struct {
	Client        *mongo.Client
	URL           string
	TestContainer testcontainers.Container
}

func NewMongoDBTestContainer(ctx context.Context, t *testing.T) (*MongoDBTestContainer, error) {
	m := &MongoDBTestContainer{}
	return m.Setup(ctx, t)
}

func (m *MongoDBTestContainer) Setup(
	ctx context.Context,
	t *testing.T,
) (*MongoDBTestContainer, error) {
	container, err := testmongodb.Run(ctx, "mongo:6")
	if err != nil {
		return nil, err
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return &MongoDBTestContainer{
		Client:        client,
		URL:           uri,
		TestContainer: container,
	}, nil
}

func (m *MongoDBTestContainer) TearDown(ctx context.Context) error {
	if m.Client != nil {
		if err := m.Client.Disconnect(ctx); err != nil {
			return err
		}
	}
	if m.TestContainer != nil {
		return m.TestContainer.Terminate(ctx)
	}
	return nil
}

func (m *MongoDBTestContainer) InsertMongoDbRecords(
	ctx context.Context,
	database, collection string,
	documents []any,
) (int, error) {
	db := m.Client.Database(database)
	col := db.Collection(collection)

	result, err := col.InsertMany(ctx, documents)
	if err != nil {
		return 0, fmt.Errorf("failed to insert mongodb records: %v", err)
	}

	return len(result.InsertedIDs), nil
}

func (m *MongoDBTestContainer) DropMongoDbCollection(
	ctx context.Context,
	database, collection string,
) error {
	db := m.Client.Database(database)
	collections, err := db.ListCollectionNames(ctx, map[string]any{"name": collection})
	if err != nil {
		return err
	}
	if len(collections) == 0 {
		return nil
	}
	return db.Collection(collection).Drop(ctx)
}
