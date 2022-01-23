package config

import log "github.com/sirupsen/logrus"

var baseLogger *log.Logger

func GetLogger() *log.Logger {
	return baseLogger
}
