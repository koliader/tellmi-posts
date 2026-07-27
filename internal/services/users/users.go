package users_service

import (
	"context"

	"github.com/jackc/pgx/v5"
	errdb "github.com/koliader/tellmi-sdk/errors/db"
	errsvc "github.com/koliader/tellmi-sdk/errors/service"
	"github.com/koliader/tellmi-sdk/rabbitmq"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	"google.golang.org/grpc/codes"
)

const userNotFound = "user not fond"

func (s *Service) CreateUser(ctx context.Context, req *rabbitmq.UserCreated) (*db.User, error) {
	user, err := s.store.CreateUser(ctx, req.Username)
	if err != nil {
		if errdb.ErrorCode(err) == errdb.UniqueViolation {
			return nil, errsvc.ErrorResponse(codes.AlreadyExists, "user with this username already exists")
		}
		return nil, errsvc.ErrorResponse(codes.Internal, "error to create user: %v", err)
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
			return nil, errsvc.ErrorResponse(codes.NotFound, userNotFound)
		}
		return nil, errsvc.ErrorResponse(codes.Internal, "error to update user: %v", err)
	}
	return &user, nil
}

func (s *Service) GetUser(ctx context.Context, username *string) (*db.User, error) {
	user, err := s.store.GetUser(ctx, *username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errsvc.ErrorResponse(codes.NotFound, userNotFound)
		}
		return nil, errsvc.ErrorResponse(codes.Internal, "error to get user: %v", err)
	}
	return &user, nil
}
