// package db

// import (
//     "context"
//     "log"
//     "time"

//     "go.mongodb.org/mongo-driver/mongo"
//     "go.mongodb.org/mongo-driver/mongo/options"
// )

// type MongoClient struct {
//     Collection *mongo.Collection
// }

// func NewMongoClient(uri, dbName string) MongoClient {
//     client, err := mongo.NewClient(options.Client().ApplyURI(uri))
//     if err != nil {
//         log.Fatalf("Mongo client creation failed: %v", err)
//     }

//     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//     defer cancel()
//     if err := client.Connect(ctx); err != nil {
//         log.Fatalf("Mongo connection failed: %v", err)
//     }

//     collection := client.Database(dbName).Collection("click_events")
//     return MongoClient{Collection: collection}
// }


package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoClient(uri, dbName string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(10)
	clientOptions.SetMinPoolSize(5)
	clientOptions.SetMaxConnIdleTime(30 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)

	// Create indexes for better performance
	if err := createIndexes(ctx, database); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return &MongoClient{
		Client:   client,
		Database: database,
	}, nil
}

func createIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("click_events")

	// Create compound index on ad_id and timestamp
	indexModel := mongo.IndexModel{
		Keys: map[string]int{
			"ad_id":     1,
			"timestamp": -1, // Descending for recent-first queries
		},
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create ad_id+timestamp index: %w", err)
	}

	// Create index on timestamp for time-based queries
	timestampIndex := mongo.IndexModel{
		Keys: map[string]int{
			"timestamp": -1,
		},
	}

	_, err = collection.Indexes().CreateOne(ctx, timestampIndex)
	if err != nil {
		return fmt.Errorf("failed to create timestamp index: %w", err)
	}

	return nil
}

func (m *MongoClient) Disconnect() {
	if m.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		m.Client.Disconnect(ctx)
	}
}