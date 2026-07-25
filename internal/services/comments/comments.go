package comments_service

import (
	"context"

	"github.com/jackc/pgx/v5"
	db_err "github.com/koliader/tellmi-posts/internal/lib/error/db"
	grpc_err "github.com/koliader/tellmi-posts/internal/lib/error/service"
	"github.com/koliader/tellmi-posts/internal/lib/token"
	pb "github.com/koliader/tellmi-posts/internal/pb"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	"google.golang.org/grpc/codes"
)

const notFound = "comment not found"

func (s *Service) CreateComment(ctx context.Context, req *pb.CreateCommentReq, payload *token.Payload) (*db.Comment, error) {
	user, err := s.users_service.GetUser(ctx, &payload.Username)
	if err != nil {
		return nil, err
	}
	arg := db.CreateCommentParams{
		Comment: req.GetComment(),
		PostID:  req.GetPostId(),
		UserID:  user.ID,
	}
	comment, err := s.store.CreateComment(ctx, arg)
	if err != nil {
		if db_err.ErrorCode(err) == db_err.ForeignKeyViolation {
			return nil, grpc_err.ErrorResponse(codes.NotFound, "invalid post or user data")
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to create comment: %v", err)
	}
	return &comment, nil
}

func (s *Service) ListCommentsByPost(ctx context.Context, req *pb.GetByIDReq) (*[]db.ListCommentsByPostRow, error) {
	comments, err := s.store.ListCommentsByPost(ctx, req.GetId())
	if err != nil {
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to list comments: %v", err)
	}
	return &comments, nil
}

func (s *Service) EditComment(ctx context.Context, req *pb.EditCommentReq, payload *token.Payload) (*db.Comment, error) {
	comment, err := s.store.GetComment(ctx, req.GetId())
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to get comment: %v", err)
	}
	if payload.Username != comment.CommenterUsername {
		return nil, grpc_err.ErrorResponse(codes.PermissionDenied, "you have no access to edit this comment")
	}
	arg := db.EditCommentParams{
		ID:      req.GetId(),
		Comment: req.GetComment(),
	}
	updatedComment, err := s.store.EditComment(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to edit comment: %v", err)
	}
	return &updatedComment, nil
}

func (s *Service) DeleteComment(ctx context.Context, req *pb.GetByIDReq, payload *token.Payload) error {
	comment, err := s.store.GetComment(ctx, req.GetId())
	if err != nil {
		if err == pgx.ErrNoRows {
			return grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return grpc_err.ErrorResponse(codes.Internal, "error to get comment: %v", err)
	}
	if payload.Username != comment.CommenterUsername {
		return grpc_err.ErrorResponse(codes.PermissionDenied, "you have no access to edit this comment")
	}
	err = s.store.DeleteComment(ctx, req.GetId())
	if err != nil {
		if err == pgx.ErrNoRows {
			return grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return grpc_err.ErrorResponse(codes.Internal, "error to delete comment: %v", err)
	}
	return nil
}
