package service

import (
	"makarov.dev/bot/internal/service/lostfilm"
	"makarov.dev/bot/internal/service/telegram"
	"makarov.dev/bot/internal/service/twitch"
)

func Init() {
	lostfilm.Init()
	go telegram.Init()
	go twitch.Init()
	go InitHealth()
}
