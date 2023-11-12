package twitch

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v2"
	"makarov.dev/bot/internal/config"
)

type ChatMessage struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	Channel      string             `bson:"channel" json:"channel"`
	User         ChatUser           `bson:"user" json:"user"`
	Message      string             `bson:"message" json:"message"`
	Raw          string             `bson:"raw" json:"raw"`
	Created      time.Time          `bson:"created" json:"created"`
	OriginalTime time.Time          `bson:"original_time" json:"originalTime"`
}

type ChatUser struct {
	Id   string `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
}

type TushqaQuote struct {
	Id      primitive.ObjectID `bson:"_id" json:"id"`
	Channel string             `bson:"channel" json:"channel"`
	Message string             `bson:"message" json:"message"`
	Created time.Time          `bson:"created" json:"created"`
}

var tushqaUserIds = make(map[string]any)

func Start(ctx context.Context) {
	log := config.GetLogger()
	cfg := config.GetConfig().Twitch
	for _, tushqaUserId := range cfg.TushqaUserIds {
		tushqaUserIds[tushqaUserId] = nil
	}
	client := twitch.NewAnonymousClient()
	log.Debug(fmt.Sprintf("Going to connect twitch channels %s", strings.Join(cfg.Channels, ", ")))
	client.Join(cfg.Channels...)
	client.OnConnect(func() {
		log.Debug("Twitch connected")
	})

	client.OnPrivateMessage(onMessageReceived)

	err := client.Connect()
	if err != nil {
		log.Error(err.Error())
		time.Sleep(10 * time.Second)
		Start(ctx)
	}

	defer func(client *twitch.Client) {
		err := client.Disconnect()
		if err != nil {
			log.Error(err.Error())
		}
	}(client)

	select {
	case <-ctx.Done():
		log.Infof("Twitch background job stopped")
		return
	default:
	}
}

func onMessageReceived(message twitch.PrivateMessage) {
	log := config.GetLogger()
	log.Trace(fmt.Sprintf(
		"Received twitch message [%s] %s: %s",
		message.Channel,
		message.User.Name,
		message.Message,
	))
	msgLink := &message
	go func() {
		err := Insert(msgLink)
		if err != nil {
			log.Error("Error while insert twitch message", err)
		}
	}()
	go func() {
		_, isTushqa := tushqaUserIds[message.User.ID]
		if !isTushqa {
			return
		}
		exists, err := TushqaQuoteExists(msgLink)
		if err != nil {
			log.Error("Error while check existed Tushqa quote", err)
			return
		}
		if exists {
			log.Trace(fmt.Sprintf("Tushqa quote %s already exists", message.Message))
			return
		}
		err = InsertTushqaQuote(msgLink)
		if err != nil {
			log.Error("Error while save Tushqa quote", err)
			return
		}
	}()
}

func Insert(m *twitch.PrivateMessage) error {
	ctx, cancel := getContext()
	defer cancel()

	_, err := getMessageCollection().InsertOne(ctx, ChatMessage{
		Id:      primitive.NewObjectID(),
		Channel: m.Channel,
		User: ChatUser{
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

func TushqaQuoteExists(m *twitch.PrivateMessage) (bool, error) {
	ctx, cancel := getContext()
	defer cancel()
	tushqaQuoteCollection := getTushqaQuoteCollection()
	limit := int64(1)
	count, err := tushqaQuoteCollection.CountDocuments(
		ctx,
		bson.M{"message": fmt.Sprintf("/^%s$/i", strings.TrimSpace(m.Message))},
		&options.CountOptions{Limit: &limit},
	)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func InsertTushqaQuote(m *twitch.PrivateMessage) error {
	ctx, cancel := getContext()
	defer cancel()
	tushqaQuoteCollection := getTushqaQuoteCollection()
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

func GetLastMessages(channel string, limit string) ([]ChatMessage, error) {
	collection := getMessageCollection()
	ctx, cancel := getContext()
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
	result := make([]ChatMessage, 0)
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetTushqaQuotes(limit string) ([]TushqaQuote, error) {
	collection := getTushqaQuoteCollection()
	ctx, cancel := getContext()
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

func getMessageCollection() *mongo.Collection {
	return config.GetDatabase().Collection("twitch_chat_messages")
}

func getTushqaQuoteCollection() *mongo.Collection {
	return config.GetDatabase().Collection("twitch_tushqa_quotes")
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
