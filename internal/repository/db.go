package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"makarov.dev/bot/internal/config"
	"time"
)

var Database *mongo.Database
var Bucket *gridfs.Bucket

func Init() {
	initDatabase()
	initBucket(Database)
}

func initDatabase() {
	cfg := config.GetConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.Uri))
	if err != nil {
		panic(err)
	}

	Database = client.Database(cfg.Database.DatabaseName)
}

func initBucket(database *mongo.Database) {
	bucket, err := gridfs.NewBucket(database)
	if err != nil {
		panic(err)
	}
	Bucket = bucket
}
