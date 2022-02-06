package lostfilm

import (
	"bytes"
	"context"
	"io"
	"time"

	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/service/telegram"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/pkg/lostfilm"
)

type Bucket interface {
	UploadFromStream(filename string, source io.Reader, opts ...*options.UploadOptions) (primitive.ObjectID, error)
}

type ServiceImpl struct {
	Client     *lostfilm.Client
	Repository repository.LostFilmRepository
	Bucket     Bucket
	Telegram   telegram.Service
}

type Service interface {
	Init()
	LastEpisodes(ctx context.Context) ([]repository.Item, error)
	StoreElement(element lostfilm.RootElement)
	Exist(page string) (bool, error)
}

func NewLostFilmService(client *lostfilm.Client, repository repository.LostFilmRepository, bucket Bucket, telegram telegram.Service) *ServiceImpl {
	return &ServiceImpl{
		Client:     client,
		Repository: repository,
		Bucket:     bucket,
		Telegram:   telegram,
	}
}

func (s *ServiceImpl) Init() {

}

func (s *ServiceImpl) LastEpisodes(ctx context.Context) ([]repository.Item, error) {
	return s.Repository.FindLatest(ctx)
}

func (s *ServiceImpl) StoreElement(element lostfilm.RootElement) {
	log := config.GetLogger()
	oldItem, err := s.Repository.GetByPage(element.Page)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Error("Error while get item by page", element.Page)
		return
	}
	if oldItem == nil {
		log.Info("Store LF item", element.Page)
	} else {
		log.Info("Try append torrent", element.Page)
	}
	episode, err := s.Client.GetEpisode(element.Page)
	if err != nil {
		log.Error("Error while get episode", err.Error())
		return
	}

	refs, err := s.Client.GetTorrentRefs(episode.Id)
	if err != nil {
		log.Error("Error while get episode refs", err.Error())
		return
	}

	nameFull := ""
	itemFiles := make([]repository.ItemFile, 0, 3)

	for _, ref := range refs {
		alreadyExist := false
		if oldItem != nil {
			for _, file := range oldItem.ItemFiles {
				if file.Quality == ref.Quality {
					alreadyExist = true
					break
				}
			}
		}
		if alreadyExist {
			continue
		}

		if nameFull == "" {
			nameFull = ref.NameFull
		}
		torrent, err := s.Client.GetTorrent(ref.TorrentUrl)
		if err != nil {
			log.Error("Error while get torrent", err.Error())
			return
		}

		objectID, err := s.Bucket.UploadFromStream(element.Name+". "+nameFull+".torrent", bytes.NewReader(torrent))
		if err != nil {
			log.Error("Error while store torrent", err.Error())
			return
		}

		itemFiles = append(itemFiles, repository.ItemFile{
			Quality:     ref.Quality,
			Description: ref.Description,
			GridFsId:    objectID,
		})
	}

	if oldItem != nil {
		oldItem.ItemFiles = append(oldItem.ItemFiles, itemFiles...)
		err := s.Repository.Update(oldItem)
		if err != nil {
			log.Error("Error while update item", oldItem.Id.Hex(), err.Error())
			return
		}
	} else {
		itemTo := &repository.Item{
			Id:              primitive.NewObjectID(),
			Page:            element.Page,
			Name:            element.Name,
			EpisodeName:     element.EpisodeName,
			EpisodeNameFull: nameFull,
			Date:            element.Date,
			Created:         time.Now(),
			ItemFiles:       itemFiles,
			Poster:          element.Poster,
		}
		err = s.Repository.Insert(itemTo)
		if err != nil {
			log.Error("Error while save item", err.Error())
			return
		}

		err = s.Telegram.SendMessageLostFilmChannel(itemTo)
		if err != nil {
			log.Error("Error while send item", err.Error())
			return
		}
	}
}

func (s *ServiceImpl) Exist(page string) (bool, error) {
	return s.Repository.Exists(page)
}
