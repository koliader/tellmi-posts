package categories_service

import (
	"context"

	"github.com/jackc/pgx/v5"
	errdb "github.com/koliader/tellmi-sdk/errors/db"
	errsvc "github.com/koliader/tellmi-sdk/errors/service"
	pb "github.com/koliader/tellmi-sdk/proto/pb"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	"google.golang.org/grpc/codes"
)

func (s *Service) CreateCategory(ctx context.Context, req *pb.CreateCategoryReq) (*db.Category, error) {
	category, err := s.store.CreateCategory(ctx, req.GetName())
	if err != nil {
		if errdb.UniqueViolation == errdb.ErrorCode(err) {
			return nil, errsvc.ErrorResponse(codes.AlreadyExists, "category with this name already exists")
		}
		return nil, errsvc.ErrorResponse(codes.Internal, "error to create category")
	}
	return &category, nil
}

func (s *Service) ListCategories(ctx context.Context) (*[]db.Category, error) {
	categories, err := s.store.ListCategories(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Internal, "error to create category")
	}
	return &categories, nil
}

func (s *Service) EditCategory(ctx context.Context, req *pb.EditCategoryReq) error {
	arg := db.EditCategoryParams{
		ID:   req.GetId(),
		Name: req.GetName(),
	}
	err := s.store.EditCategory(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errsvc.ErrorResponse(codes.NotFound, "category not found")
		}
		if errdb.UniqueViolation == errdb.ErrorCode(err) {
			return errsvc.ErrorResponse(codes.AlreadyExists, "category with this name already exists")
		}
	}
	return nil
}
