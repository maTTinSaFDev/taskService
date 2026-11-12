package service

import (
	"core/model"
	"database-service/repository"
	"errors"
	"strings"
)

type UserServiceImpl struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repository: repository}
}

func (service *UserServiceImpl) Create(user *model.User) (*model.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	user.Name = strings.TrimSpace(user.Name)
	return service.repository.Create(user)
}

func (service *UserServiceImpl) GetById(id int64) (*model.User, error) {
	if id <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	return service.repository.GetById(id)
}

func (service *UserServiceImpl) Update(user *model.User) (*model.User, error) {
	if user.Id <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	user.Name = strings.TrimSpace(user.Name)
	return service.repository.Update(user)
}

func (service *UserServiceImpl) Delete(id int64) error {
	if id <= 0 {
		return errors.New("user id must be greater than zero")
	}
	return service.repository.Delete(id)
}
