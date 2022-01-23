package twitch

import (
	"fmt"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v2"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
)

type Service struct {
	Repository *repository.TwitchChatRepository
}

func NewTwitchService(repository *repository.TwitchChatRepository) *Service {
	return &Service{Repository: repository}
}

func (s *Service) Init() {
	log := config.GetLogger()
	cfg := config.GetConfig().Twitch
	client := twitch.NewAnonymousClient()
	log.Debug("Going to connect twitch channels", strings.Join(cfg.Channels, ", "))
	client.Join(cfg.Channels...)
	client.OnConnect(func() {
		log.Debug("Twitch connected")
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		log.Trace(fmt.Sprintf(
			"Received twitch message [%s] %s: %s",
			message.Channel,
			message.User.Name,
			message.Message,
		))
		go func() {
			err := s.Repository.Insert(message)
			if err != nil {
				log.Error(err.Error())
			}
		}()
	})

	err := client.Connect()
	if err != nil {
		log.Error(err.Error())
		time.Sleep(10 * time.Second)
		s.Init()
	}

	defer func(client *twitch.Client) {
		err := client.Disconnect()
		if err != nil {
			log.Error(err.Error())
		}
	}(client)
}
