package service

import (
	"makarov.dev/bot/pkg/log"
	"time"
)

type HealthService struct {

}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s HealthService) Init() {
	for {
		log.Debug("Health ok")
		time.Sleep(1 * time.Hour)
	}
}
