package lostfilm

import (
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/pkg/lostfilm"
	"net/http"
)

type ClientService struct {
	lostfilm.Client
	HttpClient lostfilm.HttpClient
	Config     config.LostFilm
}

func (s ClientService) Init() {
	s.Client = lostfilm.Client{Config: lostfilm.ClientConfig{
		HttpClient:  s.HttpClient,
		MainPageUrl: s.Config.Domain,
		Cookie: http.Cookie{
			Name:  s.Config.CookieName,
			Value: s.Config.CookieVal,
		},
	}}
}
