package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/sirawong/point-accumulate-interview/pkg/config"
)

func NewMongoConn(cfg *config.Config) (*mongo.Database, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, nil, fmt.Errorf("could not connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, nil, fmt.Errorf("could not ping MongoDB: %w", err)
	}

	db := client.Database(cfg.MongoDBName)

	cleanup := func() {
		log.Println("Closing MongoDB connection...")
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error on disconnecting from MongoDB: %v", err)
		}
	}

	return db, cleanup, nil
}
