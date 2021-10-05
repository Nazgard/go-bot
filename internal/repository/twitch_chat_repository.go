package repository

import (
	"context"
	"github.com/gempir/go-twitch-irc/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type TwitchChatMessage struct {
	Id           primitive.ObjectID `bson:"_id"`
	Channel      string             `bson:"channel"`
	User         TwitchChatUser     `bson:"user"`
	Message      string             `bson:"message"`
	Raw          string             `bson:"raw"`
	Created      time.Time          `bson:"created"`
	OriginalTime time.Time          `bson:"original_time"`
}

type TwitchChatUser struct {
	Id   string `bson:"id"`
	Name string `bson:"name"`
}

type TwitchChatRepository struct {
	Database *mongo.Database
}

func NewTwitchChatRepository(database *mongo.Database) *TwitchChatRepository {
	return &TwitchChatRepository{Database: database}
}

func (r *TwitchChatRepository) Insert(m twitch.PrivateMessage) error {
	ctx, cancel := r.getContext()
	defer cancel()

	_, err := r.getCollection().InsertOne(ctx, TwitchChatMessage{
		Id:      primitive.NewObjectID(),
		Channel: m.Channel,
		User: TwitchChatUser{
			Id:   m.User.ID,
			Name: m.User.Name,
		},
		Message:      m.Message,
		Raw:          m.Raw,
		Created:      time.Now(),
		OriginalTime: m.Time,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *TwitchChatRepository) getCollection() *mongo.Collection {
	return r.Database.Collection("twitch_chat_messages")
}

func (r *TwitchChatRepository) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
