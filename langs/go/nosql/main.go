package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	const uri = "mongodb://localhost:27017"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Connected to MongoDB!")

	db := client.Database("testdb")
	collection := db.Collection("users")

	newUser := bson.D{
		{Key: "name", Value: "John Doe"},
		{Key: "age", Value: 24},
		{Key: "email", Value: "john@example.com"},
	}
	insertResult, err := collection.InsertOne(ctx, newUser)
	if err != nil {
		log.Fatalf("Failed to insert document: %v", err)
	}
	fmt.Printf("Inserted document with ID: %v\n", insertResult.InsertedID)

	var result bson.M
	filter := bson.D{{Key: "name", Value: "John Doe"}}
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Fatalf("Failed to find document: %v", err)
	}
	fmt.Printf("Found document: %+v\n", result)
}
