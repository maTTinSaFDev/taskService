package service

import (
	"core/model"
)

type BoardService interface {
	Create(board *model.Board) (*model.Board, error)
	GetById(id int64) (*model.Board, error)
	GetByUserId(userId int64) ([]*model.Board, error)
}
