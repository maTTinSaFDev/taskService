package service

import (
	"context"
	"core/model"
	"errors"
	"strings"
	"task-service/repository"
)

type BoardColumnServiceImpl struct {
	repository repository.BoardColumnRepository
}

func NewBoardColumnService(repository repository.BoardColumnRepository) *BoardColumnServiceImpl {
	return &BoardColumnServiceImpl{repository: repository}
}

func (service *BoardColumnServiceImpl) Create(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error) {
	if err := column.Validate(); err != nil {
		return nil, err
	}
	column.Title = strings.TrimSpace(column.Title)
	return service.repository.Create(ctx, column)
}

func (service *BoardColumnServiceImpl) Update(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error) {
	if column.Id <= 0 {
		return nil, errors.New("board column id must be greater than zero")
	}
	if err := column.Validate(); err != nil {
		return nil, err
	}
	column.Title = strings.TrimSpace(column.Title)
	return service.repository.Update(ctx, column)
}

func (service *BoardColumnServiceImpl) Reorder(ctx context.Context, boardId int64, columnIds []int64) error {
	if boardId <= 0 || len(columnIds) == 0 {
		return model.ErrInvalidBoardColumnOrder
	}
	seen := make(map[int64]struct{}, len(columnIds))
	for _, columnId := range columnIds {
		if columnId <= 0 {
			return model.ErrInvalidBoardColumnOrder
		}
		if _, exists := seen[columnId]; exists {
			return model.ErrInvalidBoardColumnOrder
		}
		seen[columnId] = struct{}{}
	}
	return service.repository.Reorder(ctx, boardId, columnIds)
}

func (service *BoardColumnServiceImpl) Delete(ctx context.Context, boardId int64, id int64, targetColumnId int64) error {
	if boardId <= 0 || id <= 0 {
		return errors.New("board and column ids must be greater than zero")
	}
	return service.repository.Delete(ctx, boardId, id, targetColumnId)
}
