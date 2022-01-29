package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
	Id      primitive.ObjectID `bson:"_id" json:"id"`
	Channel string             `bson:"channel" json:"channel"`
	Message string             `bson:"message" json:"message"`
	Created time.Time          `bson:"created" json:"created"`
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

	_, err := r.getMessageCollection().InsertOne(ctx, TwitchChatMessage{
		Id:      primitive.NewObjectID(),
		Channel: m.Channel,
		User: TwitchChatUser{
			Id:   m.User.ID,
			Name: m.User.Name,
		},
		Message:      strings.TrimSpace(m.Message),
		Raw:          m.Raw,
		Created:      time.Now(),
		OriginalTime: m.Time,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *TwitchChatRepository) TushqaQuoteExists(message string) (bool, error) {
	ctx, cancel := r.getContext()
	defer cancel()
	tushqaQuoteCollection := r.getTushqaQuoteCollection()
	limit := int64(1)
	count, err := tushqaQuoteCollection.CountDocuments(
		ctx,
		bson.M{"message": fmt.Sprintf("/^%s$/i", message)},
		&options.CountOptions{Limit: &limit},
	)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *TwitchChatRepository) InsertTushqaQuote(m twitch.PrivateMessage) error {
	ctx, cancel := r.getContext()
	defer cancel()
	tushqaQuoteCollection := r.getTushqaQuoteCollection()
	_, err := tushqaQuoteCollection.InsertOne(ctx, TushqaQuote{
		Id:      primitive.NewObjectID(),
		Channel: m.Channel,
		Message: strings.TrimSpace(m.Message),
		Created: time.Now(),
	})
	if err != nil {
		return err
	}
	return nil
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

func (r *TwitchChatRepository) GetTushqaQuotes(limit string) ([]TushqaQuote, error) {
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

func (r *TwitchChatRepository) getMessageCollection() *mongo.Collection {
	return r.Database.Collection("twitch_chat_messages")
}

func (r *TwitchChatRepository) getTushqaQuoteCollection() *mongo.Collection {
	return r.Database.Collection("twitch_tushqa_quotes")
}

func (r *TwitchChatRepository) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
