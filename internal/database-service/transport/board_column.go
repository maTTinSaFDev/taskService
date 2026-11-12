package transport

import (
	"context"
	"core/model"
	boardv1 "database-service/api/board/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (handler *BoardHandler) CreateColumn(_ context.Context, request *boardv1.CreateColumnRequest) (*boardv1.BoardColumn, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "create column request is nil")
	}
	columnModel := &model.BoardColumn{
		BoardId:  request.BoardId,
		Title:    request.Title,
		Position: request.Position,
	}
	if err := columnModel.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	column, err := handler.boardColumnService.Create(columnModel)
	if err != nil {
		return nil, boardStatusError(err)
	}
	return boardColumnToProto(column), nil
}

func (handler *BoardHandler) UpdateColumn(_ context.Context, request *boardv1.UpdateColumnRequest) (*boardv1.BoardColumn, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "update column request is nil")
	}
	columnModel := &model.BoardColumn{
		Id:       request.Id,
		BoardId:  request.BoardId,
		Title:    request.Title,
		Position: request.Position,
	}
	if columnModel.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "board column id must be greater than zero")
	}
	if err := columnModel.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	column, err := handler.boardColumnService.Update(columnModel)
	if err != nil {
		return nil, boardStatusError(err)
	}
	return boardColumnToProto(column), nil
}

func (handler *BoardHandler) ReorderColumns(_ context.Context, request *boardv1.ReorderColumnsRequest) (*emptypb.Empty, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "reorder columns request is nil")
	}
	if err := handler.boardColumnService.Reorder(request.BoardId, request.ColumnIds); err != nil {
		return nil, boardStatusError(err)
	}
	return &emptypb.Empty{}, nil
}

func (handler *BoardHandler) DeleteColumn(_ context.Context, request *boardv1.DeleteColumnRequest) (*emptypb.Empty, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "delete column request is nil")
	}
	if request.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "board column id must be greater than zero")
	}
	if err := handler.boardColumnService.Delete(request.BoardId, request.Id, request.TargetColumnId); err != nil {
		return nil, boardStatusError(err)
	}
	return &emptypb.Empty{}, nil
}
