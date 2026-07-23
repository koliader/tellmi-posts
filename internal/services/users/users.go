package users_service

import (
	"context"

	"github.com/jackc/pgx/v5"
	db_err "github.com/koliader/tellmi-posts/internal/lib/error/db"
	grpc_err "github.com/koliader/tellmi-posts/internal/lib/error/service"
	"github.com/koliader/tellmi-posts/internal/lib/rabbitmq"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	"google.golang.org/grpc/codes"
)

func (s *Service) CreateUser(ctx context.Context, req *rabbitmq.UserCreated) (*db.User, error) {
	user, err := s.store.CreateUser(ctx, req.Username)
	if err != nil {
		if db_err.ErrorCode(err) == db_err.UniqueViolation {
			return nil, grpc_err.ErrorResponse(codes.AlreadyExists, "user with this username already exists")
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to create user: %v", err)
	}
	return &user, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *rabbitmq.UserUpdated) (*db.User, error) {
	arg := db.UpdateUserParams{
		Username:    req.Username,
		NewUsername: req.NewUsername,
	}
	user, err := s.store.UpdateUser(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, grpc_err.ErrorResponse(codes.NotFound, "user not fond")
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to update user: %v", err)
	}
	return &user, nil
}
