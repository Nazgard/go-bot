package config

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sync"
	"time"
)

var onceDb sync.Once
var db *mongo.Database

func NewDatabase() *mongo.Database {
	cfg := GetConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.Uri))
	if err != nil {
		log.Fatal(err)
	}

	return client.Database(cfg.Database.DatabaseName)
}

func GetDatabase() *mongo.Database {
	if db == nil {
		onceDb.Do(func() {
			db = NewDatabase()
		})
	}
	return db
}
