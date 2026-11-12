package service

import (
	"context"
	"core/model"
)

type BoardService interface {
	Create(ctx context.Context, board *model.Board) (*model.Board, error)
	GetById(ctx context.Context, id int64) (*model.Board, error)
	GetByUserId(ctx context.Context, userId int64) ([]*model.Board, error)
}
