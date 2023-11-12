package config

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"log"
	"sync"
)

var onceBucket sync.Once
var bucket *gridfs.Bucket

func NewBucket(db *mongo.Database) *gridfs.Bucket {
	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		log.Fatal(err)
	}
	return bucket
}

func GetBucket() *gridfs.Bucket {
	if bucket == nil {
		onceBucket.Do(func() {
			bucket = NewBucket(GetDatabase())
		})
	}
	return bucket
}
