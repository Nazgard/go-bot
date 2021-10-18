package kinozal

import (
	"context"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/pkg/log"
)

var AddFavoriteCh = make(chan int64)
var DeleteFavoriteCh = make(chan int64)

type Service interface {
	IsFavorite(id int64) (bool, error)
	Exists(id int64, name string) (bool, error)
	Save(item repository.KinozalItem) error
	LastKinozalEpisodes(ctx context.Context) ([]repository.KinozalItem, error)
}

type ServiceImpl struct {
	Repository repository.KinozalRepository
}

func NewKinozalService(repository repository.KinozalRepository) *ServiceImpl {
	return &ServiceImpl{Repository: repository}
}

func (s ServiceImpl) Init() {
	go s.listenAddFavorite()
	go s.listenDeleteFavorite()
}

func (s *ServiceImpl) listenAddFavorite() {
	for id := range AddFavoriteCh {
		err := s.Repository.InsertFavorite(id)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func (s *ServiceImpl) listenDeleteFavorite() {
	for id := range DeleteFavoriteCh {
		err := s.Repository.DeleteFavorite(id)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func (s *ServiceImpl) IsFavorite(id int64) (bool, error) {
	return s.Repository.IsFavorite(id)
}

func (s *ServiceImpl) Exists(id int64, name string) (bool, error) {
	return s.Repository.Exist(id, name)
}

func (s *ServiceImpl) Save(item repository.KinozalItem) error {
	return s.Repository.Insert(item)
}

func (s *ServiceImpl) LastKinozalEpisodes(ctx context.Context) ([]repository.KinozalItem, error) {
	return s.Repository.LastEpisodes(ctx)
}
