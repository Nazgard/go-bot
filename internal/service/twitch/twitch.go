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

var tushqaUserIds = make(map[string]interface{}, 0)

func NewTwitchService(repository *repository.TwitchChatRepository) *Service {
	return &Service{Repository: repository}
}

func (s *Service) Init() {
	log := config.GetLogger()
	cfg := config.GetConfig().Twitch
	for _, tushqaUserId := range cfg.TushqaUserIds {
		tushqaUserIds[tushqaUserId] = nil
	}
	client := twitch.NewAnonymousClient()
	log.Debug(fmt.Sprintf("Going to connect twitch channels %s", strings.Join(cfg.Channels, ", ")))
	client.Join(cfg.Channels...)
	client.OnConnect(func() {
		log.Debug("Twitch connected")
	})

	client.OnPrivateMessage(s.onMessageReceived)

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

func (s *Service) onMessageReceived(message twitch.PrivateMessage) {
	log := config.GetLogger()
	log.Trace(fmt.Sprintf(
		"Received twitch message [%s] %s: %s",
		message.Channel,
		message.User.Name,
		message.Message,
	))
	msgLink := &message
	go func() {
		err := s.Repository.Insert(msgLink)
		if err != nil {
			log.Error("Error while insert twitch message", err)
		}
	}()
	go func() {
		_, isTushqa := tushqaUserIds[message.User.ID]
		if !isTushqa {
			return
		}
		exists, err := s.Repository.TushqaQuoteExists(msgLink)
		if err != nil {
			log.Error("Error while check existed Tushqa quote", err)
			return
		}
		if exists {
			log.Trace(fmt.Sprintf("Tushqa quote %s already exists", message.Message))
			return
		}
		err = s.Repository.InsertTushqaQuote(msgLink)
		if err != nil {
			log.Error("Error while save Tushqa quote", err)
			return
		}
	}()
}
