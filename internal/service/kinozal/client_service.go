package kinozal

import (
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/pkg/kinozal"
	"makarov.dev/bot/pkg/lostfilm"
)

type ClientService struct {
	kinozal.Client
	HttpClient lostfilm.HttpClient
	Config     config.Kinozal
}

func (s ClientService) Init() {
	s.Client = kinozal.Client{Config: kinozal.ClientConfig{
		HttpClient:  s.HttpClient,
		MainPageUrl: s.Config.Domain,
		Cookie:      s.Config.Cookie,
	}}
}
