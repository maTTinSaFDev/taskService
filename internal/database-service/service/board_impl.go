package service

import (
	"core/model"
	"database-service/repository"
	"errors"
	"strings"
)

type BoardServiceImpl struct {
	repository repository.BoardRepository
}

func NewBoardService(repository repository.BoardRepository) *BoardServiceImpl {
	return &BoardServiceImpl{repository: repository}
}

func (service *BoardServiceImpl) Create(board *model.Board) (*model.Board, error) {
	if err := board.Validate(); err != nil {
		return nil, err
	}
	board.Title = strings.TrimSpace(board.Title)
	return service.repository.Create(board)
}

func (service *BoardServiceImpl) GetById(id int64) (*model.Board, error) {
	if id <= 0 {
		return nil, errors.New("board id must be greater than zero")
	}
	return service.repository.GetById(id)
}

func (service *BoardServiceImpl) GetByUserId(userId int64) ([]*model.Board, error) {
	if userId <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	return service.repository.GetByUserId(userId)
}
