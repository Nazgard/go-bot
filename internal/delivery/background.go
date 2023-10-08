package delivery

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
	kinozalService "makarov.dev/bot/internal/service"
	kinozalClient "makarov.dev/bot/pkg/kinozal"
	lfClient "makarov.dev/bot/pkg/lostfilm"
)

// BackgroundJob интерфейс для всех периодических задач
type BackgroundJob interface {
	// Start блокирует текущую горутину. Выполняет какую-либо логику в заданный период
	Start()
}

// startAllBackgroundJobs запускает все периодические задачи. Не блокирует текущую горутину
func startAllBackgroundJobs() {
	kz := kinozalBackgroundJob{}
	go kz.Start()

	lf := lostFilmBackgroundJob{}
	go lf.Start()

	tg := telegramBackgroundJob{}
	go tg.Start()

	h := healthBackgroundJob{}
	go h.Start()

	t := twitchBackgroundJob{}
	go t.Start()
}

type kinozalBackgroundJob struct{}

func (c *kinozalBackgroundJob) Start() {
	log := config.GetLogger()
	if !config.GetConfig().Kinozal.Enable {
		log.Info("Kinozal integration disabled")
		return
	}
	ch := make(chan int64)
	client := kinozalClient.GetDefaultClient()
	service := kinozalService.GetKinozalService()
	bucket := repository.GetBucket()
	go client.Listing(ch, time.Minute)
	for id := range ch {
		favorite, err := service.IsFavorite(id)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if !favorite {
			log.Tracef("Kinozal item %d is not favorite", id)
			continue
		}

		name, err := client.GetName(id)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		exist, err := service.Exists(id, name)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if exist {
			log.Tracef("Kinozal item %d - %s already stored", id, name)
			continue
		}

		element, _ := client.GetElement(id)

		objectID, err := bucket.UploadFromStream(strconv.FormatInt(id, 10)+".torrent", bytes.NewReader(element.Torrent))
		if err != nil {
			log.Error("Error while store torrent", err.Error())
			continue
		}

		log.Infof("Store KZ item %s (%d)", element.Name, id)
		if config.GetConfig().Redis.Enable {
			repository.GetRedis().Del(context.Background(), "kz")
		}
		item := repository.KinozalItem{
			Id:       primitive.NewObjectID(),
			Name:     element.Name,
			DetailId: id,
			GridFsId: objectID,
			Created:  time.Now(),
		}
		err = service.Save(&item)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		service.SendToTelegram(&item)
	}
}

type lostFilmBackgroundJob struct{}

func (c *lostFilmBackgroundJob) Start() {
	log := config.GetLogger()
	if !config.GetConfig().LostFilm.Enable {
		log.Info("LostFilm integration disabled")
		return
	}
	ch := make(chan lfClient.RootElement)
	client := lfClient.GetDefaultClient()
	service := kinozalService.GetLostFilmService()
	go client.Listing(ch, time.Minute)
	for element := range ch {
		exist, err := service.Exist(element.Page)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if exist {
			continue
		}
		go service.StoreElement(element)
	}
}

type telegramBackgroundJob struct{}

func (t *telegramBackgroundJob) Start() {
	s := kinozalService.GetTelegramService()
	s.Start()
}

type healthBackgroundJob struct{}

func (h *healthBackgroundJob) Start() {
	log := config.GetLogger()
	for {
		log.Debug("Health ok")
		time.Sleep(1 * time.Hour)
	}
}

type twitchBackgroundJob struct{}

func (t *twitchBackgroundJob) Start() {
	s := kinozalService.GetTwitchService()
	s.Start()
}
