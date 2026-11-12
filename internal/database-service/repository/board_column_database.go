package repository

import (
	"context"
	"core/model"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BoardColumnRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewBoardColumnRepository(pool *pgxpool.Pool) BoardColumnRepository {
	return &BoardColumnRepositoryImpl{pool: pool}
}

func (repository *BoardColumnRepositoryImpl) Create(column *model.BoardColumn) (*model.BoardColumn, error) {
	ctx := context.Background()
	err := repository.pool.QueryRow(ctx,
		"INSERT INTO board_columns (board_id, title, position) VALUES ($1, $2, $3) RETURNING id",
		column.BoardId,
		column.Title,
		column.Position,
	).Scan(&column.Id)
	if err != nil {
		return nil, err
	}
	return column, nil
}

func (repository *BoardColumnRepositoryImpl) Update(column *model.BoardColumn) (*model.BoardColumn, error) {
	ctx := context.Background()
	err := repository.pool.QueryRow(ctx,
		"UPDATE board_columns SET title = $2, position = $3 WHERE id = $1 AND board_id = $4 RETURNING id",
		column.Id,
		column.Title,
		column.Position,
		column.BoardId,
	).Scan(&column.Id)
	if err != nil {
		return nil, mapBoardColumnLookupError(err)
	}
	return column, nil
}

func (repository *BoardColumnRepositoryImpl) Reorder(boardId int64, columnIds []int64) error {
	ctx := context.Background()
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var lockedBoardId int64
	if err := tx.QueryRow(ctx, "SELECT id FROM boards WHERE id = $1 FOR UPDATE", boardId).Scan(&lockedBoardId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrBoardNotFound
		}
		return err
	}

	rows, err := tx.Query(ctx, "SELECT id FROM board_columns WHERE board_id = $1 FOR UPDATE", boardId)
	if err != nil {
		return err
	}
	boardColumnIds := make(map[int64]struct{}, len(columnIds))
	for rows.Next() {
		var columnId int64
		if err := rows.Scan(&columnId); err != nil {
			rows.Close()
			return err
		}
		boardColumnIds[columnId] = struct{}{}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(boardColumnIds) != len(columnIds) {
		return model.ErrInvalidBoardColumnOrder
	}
	for _, columnId := range columnIds {
		if _, ok := boardColumnIds[columnId]; !ok {
			return model.ErrInvalidBoardColumnOrder
		}
	}

	commandTag, err := tx.Exec(ctx, `
		UPDATE board_columns AS board_column
		SET position = ordered.position
		FROM (
			SELECT id, ordinality - 1 AS position
			FROM unnest($2::bigint[]) WITH ORDINALITY AS input(id, ordinality)
		) AS ordered
		WHERE board_column.board_id = $1 AND board_column.id = ordered.id
	`, boardId, columnIds)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != int64(len(columnIds)) {
		return model.ErrInvalidBoardColumnOrder
	}
	return tx.Commit(ctx)
}

func (repository *BoardColumnRepositoryImpl) Delete(expectedBoardId int64, id int64, targetColumnId int64) error {
	ctx := context.Background()
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var boardId int64
	var doneColumnId int64
	err = tx.QueryRow(ctx,
		"SELECT board_columns.board_id, boards.done_column_id FROM board_columns JOIN boards ON boards.id = board_columns.board_id WHERE board_columns.id = $1",
		id,
	).Scan(&boardId, &doneColumnId)
	if err != nil {
		return mapBoardColumnLookupError(err)
	}
	if boardId != expectedBoardId {
		return model.ErrBoardColumnNotFound
	}
	if id == doneColumnId {
		return model.ErrDoneColumnCannotBeDeleted
	}

	var taskCount int64
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE column_id = $1", id).Scan(&taskCount); err != nil {
		return err
	}
	if taskCount > 0 {
		if targetColumnId <= 0 {
			return model.ErrTargetColumnRequired
		}
		var targetBoardId int64
		if err := tx.QueryRow(ctx, "SELECT board_id FROM board_columns WHERE id = $1", targetColumnId).Scan(&targetBoardId); err != nil {
			return mapBoardColumnLookupError(err)
		}
		if targetBoardId != boardId || targetColumnId == id {
			return model.ErrTargetColumnBelongsToBoard
		}
		if _, err := tx.Exec(ctx, "UPDATE tasks SET column_id = $2 WHERE column_id = $1", id, targetColumnId); err != nil {
			return err
		}
	}

	commandTag, err := tx.Exec(ctx, "DELETE FROM board_columns WHERE id = $1", id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return model.ErrBoardColumnNotFound
	}
	return tx.Commit(ctx)
}

func mapBoardColumnLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrBoardColumnNotFound
	}
	return err
}
