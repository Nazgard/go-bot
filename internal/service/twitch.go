package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gempir/go-twitch-irc/v2"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
)

type TwitchServiceImpl struct {
	Repository *repository.TwitchChatRepositoryImpl
}

var tushqaUserIds = make(map[string]any, 0)

var once = sync.Once{}
var s TwitchService

func NewTwitchService() *TwitchServiceImpl {
	return &TwitchServiceImpl{Repository: repository.GetTwitchChatRepository()}
}

func GetTwitchService() TwitchService {
	if s == nil {
		once.Do(func() {
			s = NewTwitchService()
		})
	}
	return s
}

func (s *TwitchServiceImpl) Start() {
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
		s.Start()
	}

	defer func(client *twitch.Client) {
		err := client.Disconnect()
		if err != nil {
			log.Error(err.Error())
		}
	}(client)
}

func (s *TwitchServiceImpl) onMessageReceived(message twitch.PrivateMessage) {
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
