package crawler

import (
	"makarov.dev/bot/internal/service/lostfilm"
	"makarov.dev/bot/pkg/log"
	lfClient "makarov.dev/bot/pkg/lostfilm"
	"time"
)

type LostFilmCrawler struct {
	Service lostfilm.Service
	Client  lfClient.Client
}

func (c *LostFilmCrawler) Start() {
	ch := make(chan lfClient.RootElement)
	go c.Client.Listing(ch, time.Minute)
	for element := range ch {
		exist, err := c.Service.Exist(element.Page)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if exist {
			continue
		}
		c.Service.StoreElement(element)
	}
}
