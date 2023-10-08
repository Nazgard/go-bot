package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/pkg/lostfilm"
)

type KinozalService interface {
	IsFavorite(id int64) (bool, error)
	Exists(id int64, name string) (bool, error)
	Save(item *repository.KinozalItem) error
	LastKinozalEpisodes(ctx context.Context) ([]repository.KinozalItem, error)
	AddFavorite(id int64) error
	DeleteFavorite(id int64) error
	SendToTelegram(item *repository.KinozalItem)
}

type LostFilmService interface {
	LastEpisodes(ctx context.Context) ([]repository.Item, error)
	StoreElement(element lostfilm.RootElement)
	Exist(page string) (bool, error)
}

type TelegramService interface {
	Start()
	SendMessageLostFilmChannel(lfItem *repository.Item) error
	SendMessageKinozalChannel(kzItem *repository.KinozalItem) error
}

type TwitchService interface {
	Start()
}

type FileService interface {
	GetFile(fileId *primitive.ObjectID) (*gridfs.DownloadStream, error)
	LogDownload(ctx *gin.Context, fileId primitive.ObjectID) error
}

type Bucket interface {
	OpenDownloadStream(fileID any) (*gridfs.DownloadStream, error)
	UploadFromStream(filename string, source io.Reader, opts ...*options.UploadOptions) (primitive.ObjectID, error)
}
