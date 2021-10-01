package repository

import (
	"fmt"
	"reflect"
)

type Repository interface {

}

type Registry struct {
	repositories    map[reflect.Type]Repository
	repositoryTypes []reflect.Type
}

func NewRegistry() *Registry {
	return &Registry{
		repositories:    make(map[reflect.Type]Repository),
		repositoryTypes: make([]reflect.Type, 0),
	}
}

func (s *Registry) RegisterService(service Repository) error {
	kind := reflect.TypeOf(service)
	if _, exists := s.repositories[kind]; exists {
		return fmt.Errorf("repository already exists: %v", kind)
	}
	s.repositories[kind] = service
	s.repositoryTypes = append(s.repositoryTypes, kind)
	return nil
}
