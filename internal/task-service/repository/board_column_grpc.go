package repository

import (
	"context"
	"core/model"
	boardv1 "database-service/api/board/v1"
)

type GRPCBoardColumnRepository struct {
	client boardv1.BoardServiceClient
}

func NewGRPCBoardColumnRepository(client boardv1.BoardServiceClient) *GRPCBoardColumnRepository {
	return &GRPCBoardColumnRepository{client: client}
}

func (repository *GRPCBoardColumnRepository) Create(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error) {
	createdColumn, err := repository.client.CreateColumn(ctx, &boardv1.CreateColumnRequest{
		BoardId:  column.BoardId,
		Title:    column.Title,
		Position: column.Position,
	})
	if err != nil {
		return nil, err
	}
	return boardColumnFromProto(createdColumn)
}

func (repository *GRPCBoardColumnRepository) Update(ctx context.Context, column *model.BoardColumn) (*model.BoardColumn, error) {
	updatedColumn, err := repository.client.UpdateColumn(ctx, &boardv1.UpdateColumnRequest{
		Id:       column.Id,
		BoardId:  column.BoardId,
		Title:    column.Title,
		Position: column.Position,
	})
	if err != nil {
		return nil, err
	}
	return boardColumnFromProto(updatedColumn)
}

func (repository *GRPCBoardColumnRepository) Reorder(ctx context.Context, boardId int64, columnIds []int64) error {
	_, err := repository.client.ReorderColumns(ctx, &boardv1.ReorderColumnsRequest{BoardId: boardId, ColumnIds: columnIds})
	return err
}

func (repository *GRPCBoardColumnRepository) Delete(ctx context.Context, boardId int64, id int64, targetColumnId int64) error {
	_, err := repository.client.DeleteColumn(ctx, &boardv1.DeleteColumnRequest{BoardId: boardId, Id: id, TargetColumnId: targetColumnId})
	return err
}
