package background

import (
	"context"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/integration/lostfilm"
	lfClient "makarov.dev/bot/pkg/lostfilm"
	"time"
)

type lostFilmBackgroundJob struct {
	ctx context.Context
}

func newLostFilmBackgroundJob(ctx context.Context) *lostFilmBackgroundJob {
	return &lostFilmBackgroundJob{ctx: ctx}
}

func (c *lostFilmBackgroundJob) Start() {
	log := config.GetLogger()
	if !config.GetConfig().LostFilm.Enable {
		log.Info("LostFilm integration disabled")
		return
	}
	ch := make(chan lfClient.RootElement)
	client := lfClient.GetDefaultClient()

	go client.Listing(ch, time.Minute)

	for element := range ch {
		select {
		case <-c.ctx.Done():
			log.Infof("LostFilm background job stopped")
			return
		default:
			exist, err := lostfilm.Exists(element.Page)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			if exist {
				continue
			}
			go lostfilm.StoreElement(element)
		}
	}
}
