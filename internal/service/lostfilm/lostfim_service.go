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
	item, err := s.Repository.GetByPage(element.Page)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Errorf("Error while get item by page %s", element.Page)
		return
	}
	if item == nil {
		log.Infof("Store LF item %s", element.Page)
	} else {
		log.Infof("Try append torrent %s", element.Page)
	}
	episode, err := s.Client.GetEpisode(element.Page)
	if err != nil {
		log.Errorf("Error while get episode %s", err.Error())
		return
	}

	refs, err := s.Client.GetTorrentRefs(episode.Id)
	if err != nil {
		log.Errorf("Error while get episode refs %s", err.Error())
		return
	}

	nameFull := ""
	itemFiles := make([]repository.ItemFile, 0, 3)

	for _, ref := range refs {
		alreadyExist := false
		if item != nil {
			for _, file := range item.ItemFiles {
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
			log.Errorf("Error while get torrent %s", err.Error())
			return
		}

		objectID, err := s.Bucket.UploadFromStream(element.Name+". "+nameFull+".torrent", bytes.NewReader(torrent))
		if err != nil {
			log.Errorf("Error while store torrent %s", err.Error())
			return
		}

		itemFiles = append(itemFiles, repository.ItemFile{
			Quality:     ref.Quality,
			Description: ref.Description,
			GridFsId:    objectID,
		})
	}

	if item != nil {
		item.ItemFiles = append(item.ItemFiles, itemFiles...)
		err := s.Repository.Update(item)
		if err != nil {
			log.Errorf("Error while update item %s %s", item.Id.Hex(), err.Error())
			return
		}
	} else {
		item = &repository.Item{
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
		err = s.Repository.Insert(item)
		if err != nil {
			log.Errorf("Error while save item %s", err.Error())
			return
		}
	}
	if len(itemFiles) == 3 {
		err = s.Telegram.SendMessageLostFilmChannel(item)
		if err != nil {
			log.Errorf("%s (channel id %d) %s",
				"Error while send lostfilm item to telegram channel",
				config.GetConfig().Telegram.LostFilmUpdateChannel,
				err.Error(),
			)
			return
		}
	}
}

func (s *ServiceImpl) Exist(page string) (bool, error) {
	return s.Repository.Exists(page)
}
