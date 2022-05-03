package service

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"makarov.dev/bot/internal/repository"
	"sync"
)

type FileServiceImpl struct {
	Bucket     Bucket
	Repository repository.FileRepository
}

var onceFileService = sync.Once{}
var fileService FileService

func NewFileService() *FileServiceImpl {
	return &FileServiceImpl{Bucket: repository.GetBucket(), Repository: repository.GetFileRepository()}
}

func GetFileService() FileService {
	if fileService == nil {
		onceFileService.Do(func() {
			fileService = NewFileService()
		})
	}
	return fileService
}

func (s *FileServiceImpl) Start() {

}

func (s *FileServiceImpl) GetFile(fileId *primitive.ObjectID) (*gridfs.DownloadStream, error) {
	return s.Bucket.OpenDownloadStream(fileId)
}

func (s *FileServiceImpl) LogDownload(ctx *gin.Context, fileId primitive.ObjectID) error {
	return s.Repository.Log(ctx, fileId)
}
