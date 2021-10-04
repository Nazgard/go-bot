package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"makarov.dev/bot/internal/config"
	"time"
)

func InitDatabase() *mongo.Database {
	cfg := config.GetConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.Uri))
	if err != nil {
		panic(err)
	}

	return client.Database(cfg.Database.DatabaseName)
}

func InitBucket(db *mongo.Database) *gridfs.Bucket {
	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		panic(err)
	}
	return bucket
}
