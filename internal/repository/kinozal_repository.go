package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"makarov.dev/bot/pkg/log"
	"time"
)

type KinozalRepository interface {
	IsFavorite(id int64) (bool, error)
	Exist(id int64, name string) (bool, error)
	Insert(item KinozalItem) error
	InsertFavorite(detailId int64) error
	DeleteFavorite(detailId int64) error
	LastEpisodes() ([]KinozalItem, error)
}

type Favorite struct {
	Id       primitive.ObjectID `bson:"_id"`
	DetailId int64              `bson:"detail_id"`
}

type KinozalItem struct {
	Id       primitive.ObjectID `bson:"_id"`
	Name     string             `bson:"name"`
	DetailId int64              `bson:"detail_id"`
	GridFsId primitive.ObjectID `bson:"grid_fs_id"`
	Created  time.Time          `bson:"created"`
}

type KinozalRepositoryImpl struct {
	Database *mongo.Database
}

func NewKinozalRepository(database *mongo.Database) *KinozalRepositoryImpl {
	return &KinozalRepositoryImpl{Database: database}
}

func (r *KinozalRepositoryImpl) IsFavorite(id int64) (bool, error) {
	ctx, cancelFunc := r.getContext()
	defer cancelFunc()
	result := r.getFavoriteCollection().FindOne(ctx, bson.D{{"detail_id", id}})
	if result.Err() != nil {
		if result.Err() != mongo.ErrNoDocuments {
			return false, result.Err()
		} else {
			return false, nil
		}
	}
	return true, nil
}

func (r *KinozalRepositoryImpl) Exist(id int64, name string) (bool, error) {
	ctx, cancelFunc := r.getContext()
	defer cancelFunc()
	result := r.getItemsCollection().FindOne(ctx, bson.M{"detail_id": id, "name": name})
	if result.Err() != nil {
		if result.Err() != mongo.ErrNoDocuments {
			return false, result.Err()
		} else {
			return false, nil
		}
	}
	return true, nil
}

func (r *KinozalRepositoryImpl) Insert(item KinozalItem) error {
	ctx, cancelFunc := r.getContext()
	defer cancelFunc()

	_, err := r.getItemsCollection().InsertOne(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

func (r *KinozalRepositoryImpl) InsertFavorite(detailId int64) error {
	ctx, cancelFunc := r.getContext()
	defer cancelFunc()

	_, err := r.getFavoriteCollection().InsertOne(ctx, Favorite{
		Id:       primitive.NewObjectID(),
		DetailId: detailId,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *KinozalRepositoryImpl) DeleteFavorite(detailId int64) error {
	ctx, cancelFunc := r.getContext()
	defer cancelFunc()

	_, err := r.getFavoriteCollection().DeleteOne(ctx, bson.M{"detail_id": detailId})
	if err != nil {
		return err
	}

	return nil
}

func (r *KinozalRepositoryImpl) LastEpisodes() ([]KinozalItem, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	cursor, err := r.getItemsCollection().Find(ctx, bson.D{}, &options.FindOptions{
		Sort: bson.D{{"created", -1}},
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	items := make([]KinozalItem, 0)

	for cursor.Next(ctx) {
		item := KinozalItem{}
		err := cursor.Decode(&item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *KinozalRepositoryImpl) getFavoriteCollection() *mongo.Collection {
	return r.Database.Collection("kinozal_favorites")
}

func (r *KinozalRepositoryImpl) getItemsCollection() *mongo.Collection {
	return r.Database.Collection("kinozal_items")
}

func (r *KinozalRepositoryImpl) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
