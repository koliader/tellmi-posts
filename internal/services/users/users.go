package users_service

import (
	"context"

	"github.com/koliader/tellmi-posts/internal/domain/models"
	db_err "github.com/koliader/tellmi-posts/internal/lib/error/db"
	grpc_err "github.com/koliader/tellmi-posts/internal/lib/error/service"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	"google.golang.org/grpc/codes"
)

func (s *Service) CreateUser(ctx context.Context, req *models.CreateUserReq) (*db.User, error) {
	user, err := s.store.CreateUser(ctx, req.Username)
	if err != nil {
		if db_err.ErrorCode(err) == db_err.UniqueViolation {
			return nil, grpc_err.ErrorResponse(codes.AlreadyExists, "user with this username already exists")
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to create user: %v", err)
	}
	return &user, nil
}
