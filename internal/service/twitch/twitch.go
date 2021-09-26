package twitch

import (
	"github.com/gempir/go-twitch-irc/v2"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/pkg/log"
	"strings"
)

var r *repository.TwitchChatRepository

func Init() {
	initRepository()

	cfg := config.GetConfig().Twitch
	client := twitch.NewAnonymousClient()
	log.Debug("Going to connect twitch channels", strings.Join(cfg.Channels, ", "))
	client.Join(cfg.Channels...)

	client.OnConnect(func() {
		log.Debug("Twitch connected")
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		go r.Insert(message)
	})

	client.Connect()

	defer client.Disconnect()
}

func initRepository() {
	if r == nil {
		r = &repository.TwitchChatRepository{Database: repository.Database}
	}
}
