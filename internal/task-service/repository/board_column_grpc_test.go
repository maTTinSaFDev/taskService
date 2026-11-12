package repository_test

import (
	"context"
	boardv1 "database-service/api/board/v1"
	"errors"
	"task-service/repository"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type boardColumnServiceClientStub struct {
	boardv1.BoardServiceClient
	reorderColumnsFunc func(context.Context, *boardv1.ReorderColumnsRequest, ...grpc.CallOption) (*emptypb.Empty, error)
	deleteColumnFunc   func(context.Context, *boardv1.DeleteColumnRequest, ...grpc.CallOption) (*emptypb.Empty, error)
}

func (stub *boardColumnServiceClientStub) ReorderColumns(ctx context.Context, request *boardv1.ReorderColumnsRequest, options ...grpc.CallOption) (*emptypb.Empty, error) {
	return stub.reorderColumnsFunc(ctx, request, options...)
}

func (stub *boardColumnServiceClientStub) DeleteColumn(ctx context.Context, request *boardv1.DeleteColumnRequest, options ...grpc.CallOption) (*emptypb.Empty, error) {
	return stub.deleteColumnFunc(ctx, request, options...)
}

func TestGRPCBoardColumnRepositoryReorderSendsCompleteOrder(t *testing.T) {
	client := &boardColumnServiceClientStub{
		reorderColumnsFunc: func(_ context.Context, request *boardv1.ReorderColumnsRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			if request.BoardId != 7 {
				t.Fatalf("board id = %d, want 7", request.BoardId)
			}
			want := []int64{11, 8, 9}
			if len(request.ColumnIds) != len(want) {
				t.Fatalf("column ids = %v, want %v", request.ColumnIds, want)
			}
			for index := range want {
				if request.ColumnIds[index] != want[index] {
					t.Fatalf("column ids = %v, want %v", request.ColumnIds, want)
				}
			}
			return &emptypb.Empty{}, nil
		},
	}

	if err := repository.NewGRPCBoardColumnRepository(client).Reorder(context.Background(), 7, []int64{11, 8, 9}); err != nil {
		t.Fatalf("reorder: %v", err)
	}
}

func TestGRPCBoardColumnRepositoryDeletePreservesError(t *testing.T) {
	wantErr := status.Error(codes.FailedPrecondition, "done column cannot be deleted")
	client := &boardColumnServiceClientStub{
		deleteColumnFunc: func(_ context.Context, request *boardv1.DeleteColumnRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			if request.BoardId != 7 || request.Id != 11 || request.TargetColumnId != 8 {
				t.Fatalf("request = %+v", request)
			}
			return nil, wantErr
		},
	}

	err := repository.NewGRPCBoardColumnRepository(client).Delete(context.Background(), 7, 11, 8)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}
