package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

/**
 * DB holds the active database instance.
 * Other packages receive this via function parameters, never as a global.
 */
type DB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

/** 
 * ConnectDB establishes a connection to MongoDB using the URI from Config.
 * It pings the server to confirm the connection is alive before returning.
 * Call this once in main.go - pass the result to repository functions.
*/
func ConnectDB(cfg *Config) *DB {
	// Every MongoDB operation should have a timeout. 10s is well for connection establish.
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	// Build the client options from our URI
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client:%v", err)
	}

	// Ping forces a real connection attempt
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("Failed to ping MongoDB - is it running? Error:%v", err)
	}

	log.Printf("Connected to MongoDB:%s", cfg.DBName)

	return &DB{
		Client: client,
		Database: client.Database(cfg.DBName),
	}
}

/**
 * Disconnect cleanly closes the MongoDB connection.
 * Call this in main.go with defer so it runs when the server shuts down.
*/
func (db *DB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Client.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
		return
	}

	log.Println("Disconnected from MongoDB cleanly")
}

/** 
 * GetCollection returns a handle to a named collection inside our database.
 * Repository functions call this to get the collection they need.
 * Example: db.GetCollection("users") → *mongo.Collection
*/
func (db *DB) GetCollection(name string) *mongo.Collection {
	return db.Database.Collection(name)
}
