package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to source MongoDB
	sourceClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27117"))
	if err != nil {
		log.Fatal(err)
	}
	defer sourceClient.Disconnect(ctx)

	// Connect to destination MongoDB
	destClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27217"))
	if err != nil {
		log.Fatal(err)
	}
	defer destClient.Disconnect(ctx)

	// Source collection
	sourceCollection := sourceClient.Database("data").Collection("test-sync")

	// Destination collection
	destCollection := destClient.Database("data").Collection("test-sync")

	// Query to fetch documents from source collection
	cursor, err := sourceCollection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Prepare bulk operations
	var models []mongo.WriteModel
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			log.Fatal(err)
		}

		fmt.Println()
		fmt.Println(doc)
		jsonF, _ := json.MarshalIndent(doc, "", " ")
		fmt.Printf("doc: %s \n", string(jsonF))

		// Example: Modify the map dynamically if needed
		if age, ok := doc["age"].(int32); ok {
			doc["age"] = age + 1 // Increment age as an example transformation
		}

		// Example: Add or modify fields
		doc["updated_at"] = time.Now()

		// Convert the modified map back to BSON
		// updateBson, err := bson.Marshal(doc)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// jsonF, _ = json.MarshalIndent(updateBson, "", " ")
		// fmt.Printf("updateBson: %s \n", string(jsonF))

		// Create an UpdateOneModel for upsert
		upsert := true
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.D{{Key: "_id", Value: doc["_id"]}}).
			SetUpdate(bson.D{{Key: "$set", Value: doc}}).
			SetUpsert(upsert)
		models = append(models, model)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Perform bulk upsert in destination collection
	if len(models) > 0 {
		_, err := destCollection.BulkWrite(ctx, models)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Documents have been successfully migrated, transformed, and upserted.")
}
