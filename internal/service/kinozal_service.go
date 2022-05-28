package service

import (
	"context"
	"sync"

	"makarov.dev/bot/internal/repository"
)

var onceKinozalService = sync.Once{}
var kinozalService KinozalService

type KinozalServiceImpl struct {
	Repository repository.KinozalRepository
}

func NewKinozalService() *KinozalServiceImpl {
	return &KinozalServiceImpl{Repository: repository.GetKinozalRepository()}
}

func GetKinozalService() KinozalService {
	if kinozalService == nil {
		onceKinozalService.Do(func() {
			kinozalService = NewKinozalService()
		})
	}
	return kinozalService
}

func (s *KinozalServiceImpl) IsFavorite(id int64) (bool, error) {
	return s.Repository.IsFavorite(id)
}

func (s *KinozalServiceImpl) Exists(id int64, name string) (bool, error) {
	return s.Repository.Exist(id, name)
}

func (s *KinozalServiceImpl) Save(item *repository.KinozalItem) error {
	return s.Repository.Insert(item)
}

func (s *KinozalServiceImpl) LastKinozalEpisodes(ctx context.Context) ([]repository.KinozalItem, error) {
	return s.Repository.LastEpisodes(ctx)
}

func (s *KinozalServiceImpl) AddFavorite(id int64) error {
	return s.Repository.InsertFavorite(id)
}

func (s *KinozalServiceImpl) DeleteFavorite(id int64) error {
	return s.Repository.DeleteFavorite(id)
}
