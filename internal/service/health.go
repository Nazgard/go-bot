package service

import (
	"makarov.dev/bot/pkg/log"
	"time"
)

func InitHealth() {
	for {
		log.Debug("Health ok")
		time.Sleep(1 * time.Hour)
	}
}
