package categories_service

import (
	"context"

	"github.com/jackc/pgx/v5"
	db_err "github.com/koliadertellmi-posts/internal/lib/error/db"
	grpc_err "github.com/koliadertellmi-posts/internal/lib/error/service"
	pb "github.com/koliadertellmi-posts/internal/pb"
	db "github.com/koliadertellmi-posts/internal/store/db/sqlc"
	"google.golang.org/grpc/codes"
)

func (s *Service) CreateCategory(ctx context.Context, req *pb.CreateCategoryReq) (*db.Category, error) {
	category, err := s.store.CreateCategory(ctx, req.GetName())
	if err != nil {
		if db_err.UniqueViolation == db_err.ErrorCode(err) {
			return nil, grpc_err.ErrorResponse(codes.AlreadyExists, "category with this name already exists")
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to create category")
	}
	return &category, nil
}

func (s *Service) ListCategories(ctx context.Context) (*[]db.Category, error) {
	categories, err := s.store.ListCategories(ctx)
	if err != nil {
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to create category")
	}
	return &categories, nil
}

func (s *Service) EditCategory(ctx context.Context, req *pb.EditCategoryReq) error {
	arg := db.EditCategoryParams{
		ID:   int32(req.GetId()),
		Name: req.GetName(),
	}
	err := s.store.EditCategory(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return grpc_err.ErrorResponse(codes.NotFound, "category not found")
		}
		if db_err.UniqueViolation == db_err.ErrorCode(err) {
			return grpc_err.ErrorResponse(codes.AlreadyExists, "category with this name already exists")
		}
	}
	return nil
}
