package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/gempir/go-twitch-irc/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *TwitchChatRepository) GetLastMessages(channel string, limit string) ([]TwitchChatMessage, error) {
	collection := r.getCollection()
	ctx, cancel := r.getContext()
	defer cancel()
	limitIn, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		limitIn = 100
	}
	filter := bson.D{}
	if len(channel) > 0 {
		filter = bson.D{{Key: "channel", Value: channel}}
	}
	cursor, err := collection.Find(ctx, filter, &options.FindOptions{Sort: bson.D{{Key: "_id", Value: -1}}, Limit: &limitIn})
	if err != nil {
		return nil, err
	}
	arr1 := make([]TwitchChatMessage, 0)
	err = cursor.All(ctx, &arr1)
	if err != nil {
		return nil, err
	}
	return arr1, nil
}
