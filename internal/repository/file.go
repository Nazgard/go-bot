package repository

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"makarov.dev/bot/internal/config"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type FileDownloadEntry struct {
	Id         primitive.ObjectID `bson:"_id"`
	FileId     primitive.ObjectID `bson:"file_id"`
	RemoteAddr string             `bson:"remote_addr"`
	UserAgent  string             `bson:"user_agent"`
	Created    time.Time          `bson:"created"`
}

type FileRepositoryImpl struct {
	Database *mongo.Database
}

var onceFileRepository = sync.Once{}
var fileRepository FileRepository

func NewFileRepository() *FileRepositoryImpl {
	return &FileRepositoryImpl{Database: GetDatabase()}
}

func GetFileRepository() FileRepository {
	if fileRepository == nil {
		onceFileRepository.Do(func() {
			fileRepository = NewFileRepository()
		})
	}
	return fileRepository
}

func (f *FileRepositoryImpl) Log(ctx *gin.Context, fileId primitive.ObjectID) error {
	log := config.GetLogger()
	collection := f.getCollection()
	entry := FileDownloadEntry{
		Id:         primitive.NewObjectID(),
		FileId:     fileId,
		RemoteAddr: ctx.ClientIP(),
		UserAgent:  ctx.Request.UserAgent(),
		Created:    time.Now(),
	}
	_, err := collection.InsertOne(ctx, entry)
	if err != nil {
		log.Error(fmt.Sprintf("Error while persist download log %s", entry), err.Error())
		return err
	}
	return nil
}

func (f *FileRepositoryImpl) getCollection() *mongo.Collection {
	return f.Database.Collection("file_downloads")
}
