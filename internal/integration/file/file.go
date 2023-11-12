package file

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"makarov.dev/bot/internal/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type DownloadEntry struct {
	Id         primitive.ObjectID `bson:"_id"`
	FileId     primitive.ObjectID `bson:"file_id"`
	RemoteAddr string             `bson:"remote_addr"`
	UserAgent  string             `bson:"user_agent"`
	Created    time.Time          `bson:"created"`
}

func Log(ctx *gin.Context, fileId primitive.ObjectID) error {
	log := config.GetLogger()
	collection := getCollection()
	entry := DownloadEntry{
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

func GetFile(fileId *primitive.ObjectID) (*gridfs.DownloadStream, error) {
	return config.GetBucket().OpenDownloadStream(fileId)
}

func getCollection() *mongo.Collection {
	return config.GetDatabase().Collection("file_downloads")
}
