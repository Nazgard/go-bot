package app

import (
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/crawler"
	"makarov.dev/bot/internal/delivery/web"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/internal/service"
	"makarov.dev/bot/pkg/log"
)

func Init() {
	config.Init()
	log.Init()
	repository.Init()

	service.Init()

	go crawler.Init()
	go web.Init()

	log.Debug("Application started")

	select {}
}
