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
	"net/http"
	"time"
)

func Init() {
	config.Init()
	log.Init()

	//region db
	db := repository.InitDatabase()
	bucket := repository.InitBucket(db)

	lfRepository := repository.LostFilmRepositoryImpl{Database: db}
	kzRepository := repository.KinozalRepositoryImpl{Database: db}
	twitchChatRepository := repository.TwitchChatRepository{Database: db}
	//endregion

	//region services
	lfClient := lostfilmClient.Client{
		Config: lostfilmClient.ClientConfig{
			HttpClient:  &http.Client{Timeout: 30 * time.Second},
			MainPageUrl: config.GetConfig().LostFilm.Domain,
			Cookie: http.Cookie{
				Name:  config.GetConfig().LostFilm.CookieName,
				Value: config.GetConfig().LostFilm.CookieVal,
			},
		},
	}
	kzClient := kinozalClient.Client{
		Config: kinozalClient.ClientConfig{
			HttpClient:  &http.Client{Timeout: 30 * time.Second},
			MainPageUrl: config.GetConfig().Kinozal.Domain,
			Cookie:      config.GetConfig().Kinozal.Cookie,
		},
	}
	lfService := &lostfilm.ServiceImpl{
		Client:     lfClient,
		Repository: &lfRepository,
		Bucket:     bucket,
	}
	kzService := &kinozal.ServiceImpl{Repository: &kzRepository}
	telegramService := &telegram.ServiceImpl{}
	twitchService := &twitch.Service{Repository: twitchChatRepository}
	healthService := &service.HealthService{}
	fileService := &service.FileServiceImpl{Bucket: bucket}

	go lfService.Init()
	go kzService.Init()
	go telegramService.Init()
	go twitchService.Init()
	go healthService.Init()
	go fileService.Init()
	//endregion

	//region crawlers
	lostFilmCrawler := crawler.LostFilmCrawler{Service: lfService, Client: lfClient}
	go lostFilmCrawler.Start()

	kinozalCrawler := crawler.KinozalCrawler{Service: kzService, Client: kzClient, Bucket: *bucket}
	go kinozalCrawler.Start()
	//endregion

	//region web
	wr := web.Registry{
		FileService: fileService,
		LFService:   lfService,
		KZService:   kzService,
	}
	go wr.Init()
	//endregion

	log.Debug("Application started")

	select {}
}
