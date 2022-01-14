package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"makarov.dev/bot/internal/config"
	"time"
)

func NewDatabase() *mongo.Database {
	cfg := config.GetConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.Uri))
	if err != nil {
		log.Fatal(err)
	}

	return client.Database(cfg.Database.DatabaseName)
}

func NewBucket(db *mongo.Database) *gridfs.Bucket {
	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		log.Fatal(err)
	}
	return bucket
}
