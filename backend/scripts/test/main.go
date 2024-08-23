package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Use Connect directly instead of mongo.NewClient
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27117"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	dbName := "data"
	collectionName := "test-sync"
	collection := client.Database(dbName).Collection(collectionName)

	// Create some documents with ObjectID as _id
	docs := []interface{}{
		bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "Alice"},
			{Key: "age", Value: 30},
			{Key: "minKeyField", Value: primitive.MinKey{}},
			{Key: "maxKeyField", Value: primitive.MaxKey{}},
		},
		bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "string", Value: "Hello, Alisha!"},
			{Key: "bool", Value: true},
			{Key: "int32", Value: int32(42)},
			{Key: "int64", Value: int64(92233720)},
			{Key: "double", Value: 3.14159},
			{Key: "decimal128", Value: primitive.NewDecimal128(3, 14159)},
			{Key: "date", Value: primitive.NewDateTimeFromTime(time.Now())},
			{Key: "timestamp", Value: primitive.Timestamp{T: 1645553494, I: 1}},
			{Key: "null", Value: primitive.Null{}},
			{Key: "regex", Value: primitive.Regex{Pattern: "^test", Options: "i"}},
			{Key: "array", Value: bson.A{"apple", "banana", "cherry"}},
			{Key: "embedded_document", Value: bson.D{
				{Key: "name", Value: "Alisha"},
				{Key: "age", Value: 30},
			}},
			{Key: "binary", Value: primitive.Binary{Subtype: 0x80, Data: []byte("binary data")}},
			{Key: "undefined", Value: primitive.Undefined{}},
			{Key: "object_id", Value: primitive.NewObjectID()},
			{Key: "min_key", Value: primitive.MinKey{}},
			{Key: "max_key", Value: primitive.MaxKey{}},
		},
	}

	// Insert the documents
	_, err = collection.InsertMany(ctx, docs)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare a bulk write with updates
	objectIdToUpdate := docs[0].(bson.D).Map()["_id"].(primitive.ObjectID)
	upsert := true
	models := []mongo.WriteModel{
		mongo.NewUpdateOneModel().
			SetFilter(bson.D{{Key: "_id", Value: objectIdToUpdate}}).
			SetUpdate(bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: "Alice Updated"}}}}).
			SetUpsert(upsert),
		mongo.NewUpdateOneModel().
			SetFilter(bson.D{{Key: "_id", Value: docs[1].(bson.D).Map()["_id"]}}).
			SetUpdate(bson.D{{Key: "$set", Value: bson.D{{Key: "age", Value: 26}}}}).
			SetUpsert(upsert),
	}

	res, err := collection.BulkWrite(ctx, models)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Modified count: %d, Upserted count: %d\n", res.ModifiedCount, res.UpsertedCount)
}
