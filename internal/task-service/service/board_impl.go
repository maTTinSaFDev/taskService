package service

import (
	"context"
	"core/model"
	"errors"
	"strings"
	"task-service/repository"
)

type BoardServiceImpl struct {
	repository repository.BoardRepository
}

func NewBoardService(repository repository.BoardRepository) *BoardServiceImpl {
	return &BoardServiceImpl{repository: repository}
}

func (service *BoardServiceImpl) Create(ctx context.Context, board *model.Board) (*model.Board, error) {
	if err := board.Validate(); err != nil {
		return nil, err
	}
	board.Title = strings.TrimSpace(board.Title)
	return service.repository.Create(ctx, board)
}

func (service *BoardServiceImpl) GetById(ctx context.Context, id int64) (*model.Board, error) {
	if id <= 0 {
		return nil, errors.New("board id must be greater than zero")
	}
	return service.repository.GetById(ctx, id)
}

func (service *BoardServiceImpl) GetByUserId(ctx context.Context, userId int64) ([]*model.Board, error) {
	if userId <= 0 {
		return nil, errors.New("user id must be greater than zero")
	}
	return service.repository.GetByUserId(ctx, userId)
}
