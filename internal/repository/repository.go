package repository

import (
	"context"
	"github.com/gempir/go-twitch-irc/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileRepository interface {
	Log(ctx *gin.Context, fileId primitive.ObjectID) error
}

type KinozalRepository interface {
	IsFavorite(id int64) (bool, error)
	Exist(id int64, name string) (bool, error)
	Insert(item *KinozalItem) error
	InsertFavorite(detailId int64) error
	DeleteFavorite(detailId int64) error
	LastEpisodes(ctx context.Context) ([]KinozalItem, error)
}

type LostFilmRepository interface {
	FindLatest(ctx context.Context) ([]Item, error)
	Exists(page string) (bool, error)
	Insert(item *Item) error
	Update(item *Item) error
	GetByPage(page string) (*Item, error)
}

type TwitchChatRepository interface {
	Insert(m *twitch.PrivateMessage) error
	TushqaQuoteExists(m *twitch.PrivateMessage) (bool, error)
	InsertTushqaQuote(m *twitch.PrivateMessage) error
	GetLastMessages(channel string, limit string) ([]TwitchChatMessage, error)
	GetTushqaQuotes(limit string) ([]TushqaQuote, error)
}
