package service

import (
	"core/model"
)

type BoardColumnService interface {
	Create(column *model.BoardColumn) (*model.BoardColumn, error)
	Update(column *model.BoardColumn) (*model.BoardColumn, error)
	Reorder(boardId int64, columnIds []int64) error
	Delete(boardId int64, id int64, targetColumnId int64) error
}
