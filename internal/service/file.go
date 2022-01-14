package service

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"makarov.dev/bot/internal/repository"
)

type FileService interface {
	GetFile(fileId *primitive.ObjectID) (*gridfs.DownloadStream, error)
	LogDownload(ctx *gin.Context, fileId primitive.ObjectID) error
}

type Bucket interface {
	OpenDownloadStream(fileID interface{}) (*gridfs.DownloadStream, error)
}

type FileServiceImpl struct {
	Bucket     Bucket
	Repository repository.FileRepository
}

func NewFileService(bucket Bucket, repository repository.FileRepository) *FileServiceImpl {
	return &FileServiceImpl{Bucket: bucket, Repository: repository}
}

func (s *FileServiceImpl) Init() {

}

func (s *FileServiceImpl) GetFile(fileId *primitive.ObjectID) (*gridfs.DownloadStream, error) {
	return s.Bucket.OpenDownloadStream(fileId)
}

func (s *FileServiceImpl) LogDownload(ctx *gin.Context, fileId primitive.ObjectID) error {
	return s.Repository.Log(ctx, fileId)
}
