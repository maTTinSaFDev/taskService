package repository

import (
	"context"
	"core/model"
)

type BoardColumnRepository interface {
	Create(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error)
	Update(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error)
	Reorder(ctx context.Context, boardId int64, columnIds []int64) error
	Delete(ctx context.Context, boardId int64, id int64, targetColumnId int64) error
}
