package app

import (
	log "github.com/sirupsen/logrus"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/delivery"
)

func Init() {
	logger := log.New()
	config.Init(logger)

	delivery.StartAllInputs()

	log.Debug("Application started")

	select {}
}
