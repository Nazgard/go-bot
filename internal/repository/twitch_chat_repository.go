package repository

import (
	"context"
	log "github.com/sirupsen/logrus"
	"makarov.dev/bot/internal/config"
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

	message := strings.TrimSpace(m.Message)
	_, err := r.getMessageCollection().InsertOne(ctx, TwitchChatMessage{
		Id:      primitive.NewObjectID(),
		Channel: m.Channel,
		User: TwitchChatUser{
			Id:   m.User.ID,
			Name: m.User.Name,
		},
		Message:      message,
		Raw:          m.Raw,
		Created:      time.Now(),
		OriginalTime: m.Time,
	})
	if err != nil {
		return err
	}

	go func() {
		_, isTushqaUser := tushqaUserIds[m.User.ID]
		if isTushqaUser {
			tushqaQuoteCollection := r.getTushqaQuoteCollection()
			limit := int64(1)
			count, err := tushqaQuoteCollection.CountDocuments(
				ctx,
				bson.M{"$regex": primitive.Regex{Pattern: message, Options: "i"}},
				&options.CountOptions{Limit: &limit},
			)
			if err != nil {
				log.Error("Error while check existed Tushqa quote", err)
				return
			}
			if count > 0 {
				return
			}
			_, err = tushqaQuoteCollection.InsertOne(ctx, TushqaQuote{
				Id:      primitive.NewObjectID(),
				Channel: m.Channel,
				Message: message,
				Created: time.Now(),
			})
			if err != nil {
				log.Error("Error while save Tushqa quote", err)
			}
		}
	}()

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
