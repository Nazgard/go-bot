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

type Item struct {
	Id              primitive.ObjectID `bson:"_id"`
	Page            string             `bson:"page"`
	Name            string             `bson:"name"`
	EpisodeName     string             `bson:"episode_name"`
	EpisodeNameFull string             `bson:"episode_name_full"`
	Date            time.Time          `bson:"date"`
	Created         time.Time          `bson:"created"`
	ItemFiles       []ItemFile         `bson:"item_files"`
}

type ItemFile struct {
	Quality     string             `bson:"quality"`
	Description string             `bson:"description"`
	GridFsId    primitive.ObjectID `bson:"grid_fs_id"`
}

type LostFilmRepositoryImpl struct {
	Database *mongo.Database
}

type LostFilmRepository interface {
	FindLatest() ([]Item, error)
	Exists(page string) (bool, error)
	Insert(item *Item) error
	Update(item *Item) error
	GetByPage(page string) (*Item, error)
}

func NewLostFilmRepository(database *mongo.Database) *LostFilmRepositoryImpl {
	return &LostFilmRepositoryImpl{Database: database}
}

func (r *LostFilmRepositoryImpl) FindLatest() ([]Item, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	limit := int64(50)
	cursor, err := r.getCollection().Find(ctx, bson.D{}, &options.FindOptions{
		Sort: bson.D{{"date", -1}, {"created", -1}},
		Limit: &limit,
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	items := make([]Item, 0)

	for cursor.Next(ctx) {
		item := Item{}
		err := cursor.Decode(&item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *LostFilmRepositoryImpl) Exists(page string) (bool, error) {
	item, err := r.GetByPage(page)
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}
	return len(item.ItemFiles) >= 3, nil
}

func (r *LostFilmRepositoryImpl) Insert(item *Item) error {
	ctx, cancel := r.getContext()
	defer cancel()

	_, err := r.getCollection().InsertOne(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

func (r *LostFilmRepositoryImpl) Update(item *Item) error {
	ctx, cancel := r.getContext()
	defer cancel()

	_, err := r.getCollection().UpdateOne(ctx, bson.D{{"_id", item.Id}}, bson.M{"$set": item})
	if err != nil {
		return err
	}

	return nil
}

func (r *LostFilmRepositoryImpl) GetByPage(page string) (*Item, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	result := r.getCollection().FindOne(ctx, bson.D{{"page", page}})
	if result.Err() != nil {
		if result.Err() != mongo.ErrNoDocuments {
			return nil, result.Err()
		} else {
			return nil, nil
		}
	}
	item := Item{}
	err := result.Decode(&item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *LostFilmRepositoryImpl) getCollection() *mongo.Collection {
	return r.Database.Collection("lostfilm_items")
}

func (r *LostFilmRepositoryImpl) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
