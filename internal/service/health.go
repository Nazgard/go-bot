package service

import (
	"makarov.dev/bot/internal/config"
	"time"
)

type HealthService struct {
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Init() {
	log := config.GetLogger()
	for {
		log.Debug("Health ok")
		time.Sleep(1 * time.Hour)
	}
}
