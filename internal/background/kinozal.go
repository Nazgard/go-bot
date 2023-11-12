package background

import (
	"bytes"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/integration/kinozal"
	"makarov.dev/bot/internal/integration/telegram"
	kinozalClient "makarov.dev/bot/pkg/kinozal"
	"strconv"
	"strings"
	"time"
)

type kinozalBackgroundJob struct {
	ctx context.Context
}

func newKinozalBackgroundJob(ctx context.Context) *kinozalBackgroundJob {
	return &kinozalBackgroundJob{ctx: ctx}
}

func (c *kinozalBackgroundJob) Start() {
	log := config.GetLogger()
	if !config.GetConfig().Kinozal.Enable {
		log.Info("Kinozal integration disabled")
		return
	}

	addTelegramCmd()

	ch := make(chan int64)
	client := kinozalClient.GetDefaultClient()
	bucket := config.GetBucket()

	go client.Listing(ch, time.Minute)

	for id := range ch {
		select {
		case <-c.ctx.Done():
			log.Infof("Kinozal background job stopped")
			return
		default:
			favorite, err := kinozal.IsFavorite(id)
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
			exist, err := kinozal.Exist(id, name)
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
				config.GetRedis().Del(context.Background(), "kz")
			}
			item := kinozal.Item{
				Id:       primitive.NewObjectID(),
				Name:     element.Name,
				DetailId: id,
				GridFsId: objectID,
				Created:  time.Now(),
			}
			err = kinozal.Insert(&item)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			kinozal.SendToTelegram(&item)
		}

	}
}

func addTelegramCmd() {
	err := telegram.AddRouterFunc("/add", func(txt string) string {
		if strings.Contains(txt, "kinozal") {
			txt := strings.ReplaceAll(txt, "kinozal", "")
			txt = strings.TrimSpace(txt)
			id, err := strconv.ParseInt(txt, 10, 64)
			if err != nil {
				errMsg := fmt.Sprintf("add cmd wrong id %s", txt)
				config.GetLogger().Error(errMsg)
				return errMsg
			}
			err = kinozal.InsertFavorite(id)
			if err != nil {
				return err.Error()
			}
			return "Ok"
		}

		return "wrong add provider"
	})
	if err != nil {
		config.GetLogger().Errorf("Error while add telegram Add cmd %s", err.Error())
	}

	err = telegram.AddRouterFunc("/delete", func(txt string) string {
		if strings.Contains(txt, "kinozal") {
			txt := strings.ReplaceAll(txt, "kinozal", "")
			txt = strings.TrimSpace(txt)
			id, err := strconv.ParseInt(txt, 10, 64)
			if err != nil {
				errMsg := fmt.Sprintf("delete cmd wrong id %s", txt)
				config.GetLogger().Error(errMsg)
				return errMsg
			}
			err = kinozal.DeleteFavorite(id)
			if err != nil {
				return err.Error()
			}
			return "Ok"
		}

		return "wrong delete provider"
	})
	if err != nil {
		config.GetLogger().Errorf("Error while add telegram Delete cmd %s", err.Error())
	}
}
