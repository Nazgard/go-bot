package repository

import (
	"context"
	"makarov.dev/bot/internal/config"
	"strconv"
	"time"

	"github.com/gempir/go-twitch-irc/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TwitchChatMessage struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	Channel      string             `bson:"channel" json:"channel"`
	User         TwitchChatUser     `bson:"user" json:"user"`
	Message      string             `bson:"message" json:"message"`
	Raw          string             `bson:"raw" json:"raw"`
	Created      time.Time          `bson:"created" json:"created"`
	OriginalTime time.Time          `bson:"original_time" json:"originalTime"`
}

type TwitchChatUser struct {
	Id   string `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
}

type TushqaQuote struct {
	Id      primitive.ObjectID `bson:"id" json:"id"`
	Channel string             `bson:"channel" json:"channel"`
	Message string             `bson:"message" json:"message"`
	Created time.Time          `bson:"created" json:"created"`
}

type TwitchChatRepository struct {
	Database *mongo.Database
}

var tushqaUserIds = make(map[string]interface{}, 0)

func NewTwitchChatRepository(database *mongo.Database) *TwitchChatRepository {
	for _, tushqaUserId := range config.GetConfig().Twitch.TushqaUserIds {
		tushqaUserIds[tushqaUserId] = nil
	}
	return &TwitchChatRepository{Database: database}
}

func (r *TwitchChatRepository) Insert(m twitch.PrivateMessage) error {
	ctx, cancel := r.getContext()
	defer cancel()

	_, err := r.getMessageCollection().InsertOne(ctx, TwitchChatMessage{
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

	_, isTushqaUser := tushqaUserIds[m.User.ID]
	if isTushqaUser {
		_, err := r.getTushqaQuoteCollection().InsertOne(ctx, TushqaQuote{
			Id:      primitive.NewObjectID(),
			Channel: m.Channel,
			Message: m.Message,
			Created: time.Now(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TwitchChatRepository) getMessageCollection() *mongo.Collection {
	return r.Database.Collection("twitch_chat_messages")
}

func (r *TwitchChatRepository) GetLastMessages(channel string, limit string) ([]TwitchChatMessage, error) {
	collection := r.getMessageCollection()
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
	result := make([]TwitchChatMessage, 0)
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r TwitchChatRepository) GetTushqaQuotes(limit string) ([]TushqaQuote, error) {
	collection := r.getTushqaQuoteCollection()
	ctx, cancel := r.getContext()
	defer cancel()
	limitIn, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		limitIn = 100
	}
	filter := bson.D{}
	cursor, err := collection.Find(
		ctx,
		filter,
		&options.FindOptions{Sort: bson.D{{Key: "_id", Value: -1}}, Limit: &limitIn},
	)
	if err != nil {
		return nil, err
	}
	result := make([]TushqaQuote, 0)
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *TwitchChatRepository) getTushqaQuoteCollection() *mongo.Collection {
	return r.Database.Collection("twitch_tushqa_quotes")
}

func (r *TwitchChatRepository) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
