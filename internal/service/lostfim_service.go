package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mattn/go-mastodon"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/pkg/lostfilm"
)

type LostFilmServiceImpl struct {
	Client     *lostfilm.Client
	Repository repository.LostFilmRepository
	Bucket     Bucket
	Telegram   TelegramService
}

var onceLostFilmService = sync.Once{}
var lostFilmService LostFilmService

func NewLostFilmService(client *lostfilm.Client) *LostFilmServiceImpl {
	return &LostFilmServiceImpl{
		Client:     client,
		Repository: repository.GetLostFilmRepository(),
		Bucket:     repository.GetBucket(),
		Telegram:   GetTelegramService(),
	}
}

func GetLostFilmService() LostFilmService {
	if lostFilmService == nil {
		onceLostFilmService.Do(func() {
			lostFilmService = NewLostFilmService(lostfilm.GetDefaultClient())
		})
	}
	return lostFilmService
}

func (s *LostFilmServiceImpl) LastEpisodes(ctx context.Context) ([]repository.Item, error) {
	return s.Repository.FindLatest(ctx)
}

func (s *LostFilmServiceImpl) StoreElement(element lostfilm.RootElement) {
	cfg := config.GetConfig()
	lfCfg := cfg.LostFilm
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
	if strings.HasPrefix(element.Page, "/movies") {
		nameFull = "Фильм"
	}
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

		if cfg.Redis.Enable {
			repository.GetRedis().Del(context.Background(), "lf-"+ref.Quality)
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
		item.RetryCount++
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
	if len(item.ItemFiles) == 3 || (len(item.ItemFiles) > 0 && item.RetryCount >= lfCfg.MaxRetries) {
		go s.sendToTelegram(item)
		go sendToMastodon(*item)
	}
}

func (s *LostFilmServiceImpl) Exist(page string) (bool, error) {
	return s.Repository.Exists(page)
}

func (s *LostFilmServiceImpl) sendToTelegram(item *repository.Item) {
	err := s.Telegram.SendMessageLostFilmChannel(item)
	if err != nil {
		config.GetLogger().Errorf("%s (channel id %d) %s",
			"Error while send lostfilm item to telegram channel",
			config.GetConfig().Telegram.LostFilmUpdateChannel,
			err.Error(),
		)
		return
	}
}

func sendToMastodon(item repository.Item) {
	if len(item.ItemFiles) <= 0 {
		return
	}
	client := GetMastodonClient()
	posterRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https:%s", item.Poster), nil)
	if err != nil {
		config.GetLogger().Errorf("Error while download poster for Mastodon status %s", err.Error())
		return
	}
	response, err := client.Client.Do(posterRequest)
	if err != nil {
		config.GetLogger().Errorf("Error while download poster for Mastodon status %s", err.Error())
		return
	}
	defer response.Body.Close()
	imgBytes, err := io.ReadAll(response.Body)
	if err != nil {
		config.GetLogger().Errorf("Error while download poster for Mastodon status %s", err.Error())
		return
	}
	attachment, err := client.UploadMediaFromBytes(context.Background(), imgBytes)
	if err != nil {
		config.GetLogger().Errorf("Error while upload media for Mastodon status %s", err.Error())
		return
	}

	status := fmt.Sprintf("%s. %s\n", item.Name, item.EpisodeNameFull)
	domain := config.GetConfig().Web.Domain
	for _, f := range item.ItemFiles {
		status += fmt.Sprintf("\n%s %s/dl/%s", f.Quality, domain, f.GridFsId.Hex())
	}
	_, err = client.PostStatus(context.Background(), &mastodon.Toot{
		Status:     status,
		Visibility: mastodon.VisibilityPublic,
		Language:   "ru",
		MediaIDs:   []mastodon.ID{attachment.ID},
	})
	if err != nil {
		config.GetLogger().Errorf("Error while send Mastodon status %s", err.Error())
	} else {
		config.GetLogger().Debugf("Posted status to Mastodon for item %s", item.Page)
	}
}
