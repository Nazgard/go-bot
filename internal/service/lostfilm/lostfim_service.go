package lostfilm

import (
	"bytes"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/delivery/channels"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/pkg/log"
	"makarov.dev/bot/pkg/lostfilm"
	"net/http"
	"time"
)

var r *repository.LostFilmRepository
var Client *lostfilm.Client

func Init() {
	initClient()
	getRepository()

	go listen()
}

func storeElement(element lostfilm.RootElement) {
	oldItem, err := r.GetByPage(element.Page)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Error("Error while get item by page", element.Page)
		return
	}
	if oldItem == nil {
		log.Info("Store LF item", element.Page)
	} else {
		log.Info("Try append torrent", element.Page)
	}
	episode, err := Client.GetEpisode(element.Page)
	if err != nil {
		log.Error("Error while get episode", err.Error())
		return
	}

	refs, err := Client.GetTorrentRefs(episode.Id)
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
		torrent, err := Client.GetTorrent(ref.TorrentUrl)
		if err != nil {
			log.Error("Error while get torrent", err.Error())
			return
		}

		objectID, err := repository.Bucket.UploadFromStream("file", bytes.NewReader(torrent))
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
		err := r.Update(oldItem)
		if err != nil {
			log.Error("Error while update item", oldItem.Id.Hex(), err.Error())
			return
		}
	} else {
		err = r.Insert(&repository.Item{
			Id:              primitive.NewObjectID(),
			Page:            element.Page,
			Name:            element.Name,
			EpisodeName:     element.EpisodeName,
			EpisodeNameFull: nameFull,
			Date:            element.Date,
			Created:         time.Now(),
			ItemFiles:       itemFiles,
		})
		if err != nil {
			log.Error("Error while save item", err.Error())
			return
		}
	}
}

func getRepository() {
	if r == nil {
		r = &repository.LostFilmRepository{Database: repository.Database}
	}
}

func initClient() {
	cfg := config.GetConfig().LostFilm
	Client = &lostfilm.Client{
		Config: lostfilm.ClientConfig{
			HttpClient:  &http.Client{Timeout: 30 * time.Second},
			MainPageUrl: cfg.Domain,
			Cookie: http.Cookie{
				Name:  cfg.CookieName,
				Value: cfg.CookieVal,
			}}}
}

func listen() {
	for element := range channels.LostFilmInputChannel {
		exists, err := r.Exists(element.Page)
		if err != nil {
			panic(err)
		}
		if exists {
			continue
		}

		go storeElement(element)
	}
}

func LastEpisodes() ([]repository.Item, error) {
	return r.FindLatest()
}
