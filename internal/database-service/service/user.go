package service

import (
	"core/model"
)

type UserService interface {
	Create(user *model.User) (*model.User, error)
	GetById(id int64) (*model.User, error)
	Update(user *model.User) (*model.User, error)
	Delete(id int64) error
}
