package twitch

import (
	"github.com/gempir/go-twitch-irc/v2"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/pkg/log"
	"strings"
)

type Service struct {
	Repository *repository.TwitchChatRepository
}

func NewTwitchService(repository *repository.TwitchChatRepository) *Service {
	return &Service{Repository: repository}
}

func (s Service) Init() {
	cfg := config.GetConfig().Twitch
	client := twitch.NewAnonymousClient()
	log.Debug("Going to connect twitch channels", strings.Join(cfg.Channels, ", "))
	client.Join(cfg.Channels...)

	client.OnConnect(func() {
		log.Debug("Twitch connected")
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		go s.Repository.Insert(message)
	})

	client.Connect()

	defer client.Disconnect()
}
