package config

import log "github.com/sirupsen/logrus"

var baseLogger *log.Logger

func GetLogger() *log.Logger {
	if baseLogger == nil {
		return log.StandardLogger()
	}
	return baseLogger
}
