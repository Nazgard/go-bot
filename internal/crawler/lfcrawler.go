package crawler

import (
	"makarov.dev/bot/internal/delivery/channels"
	"makarov.dev/bot/internal/service/lostfilm"
	"time"
)

func Init() {
	go initLostFilmCrawler()
}

func initLostFilmCrawler() {
	lostfilm.Client.Listing(channels.LostFilmInputChannel, 1*time.Minute)
}
