package config

import (
	"github.com/Nazgard/logruzio"
	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

var baseLogger *log.Logger

func GetLogger() *log.Logger {
	if baseLogger == nil {
		return log.StandardLogger()
	}
	return baseLogger
}

func initLogger(logger *log.Logger) {
	baseLogger = logger
	logLevel, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	baseLogger.SetLevel(logLevel)
	logger.SetFormatter(&nested.Formatter{})
	if !config.Debug {
		hook, err := logruzio.New(config.Logzio.Host, config.Logzio.Token, "Bot", log.Fields{})
		if err != nil {
			log.Fatal(err)
		}
		logger.AddHook(hook)
	}
}
