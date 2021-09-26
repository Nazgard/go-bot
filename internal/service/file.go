package service

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"makarov.dev/bot/internal/repository"
)

func GetFile(fileId *primitive.ObjectID) (*gridfs.DownloadStream, error) {
	return repository.Bucket.OpenDownloadStream(fileId)
}
