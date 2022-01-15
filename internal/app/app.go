package app

import (
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/crawler"
	"makarov.dev/bot/internal/delivery/web"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/internal/service"
	"makarov.dev/bot/internal/service/kinozal"
	"makarov.dev/bot/internal/service/lostfilm"
	"makarov.dev/bot/internal/service/telegram"
	"makarov.dev/bot/internal/service/twitch"
	kinozalClient "makarov.dev/bot/pkg/kinozal"
	"makarov.dev/bot/pkg/log"
	lostfilmClient "makarov.dev/bot/pkg/lostfilm"
)

func Init() {
	config.Init()
	log.Init()

	//region db
	db := repository.NewDatabase()
	bucket := repository.NewBucket(db)

	lfRepository := repository.NewLostFilmRepository(db)
	kzRepository := repository.NewKinozalRepository(db)
	twitchChatRepository := repository.NewTwitchChatRepository(db)
	fileRepository := repository.NewFileRepository(db)
	//endregion

	//region services
	lfCfg := config.GetConfig().LostFilm
	lfClient := lostfilmClient.NewClient(lfCfg.CookieName, lfCfg.CookieVal)
	kzClient := kinozalClient.NewClient(config.GetConfig().Kinozal.Cookie)
	lfService := lostfilm.NewLostFilmService(lfClient, lfRepository, bucket)
	kzService := kinozal.NewKinozalService(kzRepository)
	telegramService := telegram.NewTelegramService()
	twitchService := twitch.NewTwitchService(twitchChatRepository)
	healthService := service.NewHealthService()
	fileService := service.NewFileService(bucket, fileRepository)

	go lfService.Init()
	go kzService.Init()
	go telegramService.Init()
	go twitchService.Init()
	go healthService.Init()
	go fileService.Init()
	//endregion

	//region crawlers
	go crawler.NewLostFilmCrawler(lfService, lfClient).Start()
	go crawler.NewKinozalCrawler(kzService, kzClient, bucket).Start()
	//endregion

	//region web
	wr := web.Registry{
		FileService:     fileService,
		LFService:       lfService,
		KZService:       kzService,
		TwichRepository: twitchChatRepository,
	}
	go wr.Init()
	//endregion

	log.Debug("Application started")

	select {}
}
