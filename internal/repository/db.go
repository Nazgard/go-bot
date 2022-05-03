package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"makarov.dev/bot/internal/config"
	"sync"
	"time"
)

var onceDb sync.Once
var onceBucket sync.Once
var db *mongo.Database
var bucket *gridfs.Bucket

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

func GetDatabase() *mongo.Database {
	if db == nil {
		onceDb.Do(func() {
			db = NewDatabase()
		})
	}
	return db
}

func GetBucket() *gridfs.Bucket {
	if bucket == nil {
		onceBucket.Do(func() {
			bucket = NewBucket(GetDatabase())
		})
	}
	return bucket
}
