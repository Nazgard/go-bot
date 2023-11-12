package kinozal

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/integration/telegram"
	"time"
)

type Favorite struct {
	Id       primitive.ObjectID `bson:"_id"`
	DetailId int64              `bson:"detail_id"`
}

type Item struct {
	Id       primitive.ObjectID `bson:"_id"`
	Name     string             `bson:"name"`
	DetailId int64              `bson:"detail_id"`
	GridFsId primitive.ObjectID `bson:"grid_fs_id"`
	Created  time.Time          `bson:"created"`
}

func SendToTelegram(item *Item) {
	cfg := config.GetConfig()
	_, err := telegram.SendMessage(tgbotapi.NewMessage(cfg.Telegram.KinozalUpdateChannel,
		fmt.Sprintf("Вышла новая серия - %s", item.Name)))
	if err != nil {
		config.GetLogger().Errorf("%s (channel id %d) %s",
			"Error while send kinozal item to telegram channel",
			config.GetConfig().Telegram.KinozalUpdateChannel,
			err.Error(),
		)
		return
	}
}

func IsFavorite(id int64) (bool, error) {
	ctx, cancelFunc := getContext()
	defer cancelFunc()
	result := getFavoriteCollection().FindOne(ctx, bson.D{{"detail_id", id}})
	if result.Err() != nil {
		if !errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return false, result.Err()
		} else {
			return false, nil
		}
	}
	return true, nil
}

func Exist(id int64, name string) (bool, error) {
	ctx, cancelFunc := getContext()
	defer cancelFunc()
	result := getItemsCollection().FindOne(ctx, bson.M{"detail_id": id, "name": name})
	if result.Err() != nil {
		if !errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return false, result.Err()
		} else {
			return false, nil
		}
	}
	return true, nil
}

func Insert(item *Item) error {
	ctx, cancelFunc := getContext()
	defer cancelFunc()

	_, err := getItemsCollection().InsertOne(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

func InsertFavorite(detailId int64) error {
	ctx, cancelFunc := getContext()
	defer cancelFunc()

	_, err := getFavoriteCollection().InsertOne(ctx, Favorite{
		Id:       primitive.NewObjectID(),
		DetailId: detailId,
	})
	if err != nil {
		return err
	}

	return nil
}

func DeleteFavorite(detailId int64) error {
	ctx, cancelFunc := getContext()
	defer cancelFunc()

	_, err := getFavoriteCollection().DeleteOne(ctx, bson.M{"detail_id": detailId})
	if err != nil {
		return err
	}

	return nil
}

func LastEpisodes(ctx context.Context) ([]Item, error) {
	log := config.GetLogger()
	limit := int64(50)
	cursor, err := getItemsCollection().Find(ctx, bson.D{}, &options.FindOptions{
		Sort:  bson.D{{"created", -1}},
		Limit: &limit,
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	items := make([]Item, 0)
	err = cursor.All(ctx, &items)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return items, nil
}

func getFavoriteCollection() *mongo.Collection {
	return config.GetDatabase().Collection("kinozal_favorites")
}

func getItemsCollection() *mongo.Collection {
	return config.GetDatabase().Collection("kinozal_items")
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
