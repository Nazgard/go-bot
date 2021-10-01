package service

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

type FileService interface {
	GetFile(fileId *primitive.ObjectID) (*gridfs.DownloadStream, error)
}

type Bucket interface {
	OpenDownloadStream(fileID interface{}) (*gridfs.DownloadStream, error)
}

type FileServiceImpl struct {
	Bucket Bucket
}

func (receiver FileServiceImpl) Init() {

}

func (s *FileServiceImpl) GetFile(fileId *primitive.ObjectID) (*gridfs.DownloadStream, error) {
	return s.Bucket.OpenDownloadStream(fileId)
}
