package comments_service

import (
	"context"

	"github.com/jackc/pgx/v5"
	grpc_err "github.com/koliader/tellmi-posts/internal/lib/error/service"
	pb "github.com/koliader/tellmi-posts/internal/pb"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	"google.golang.org/grpc/codes"
)

const notFound = "comment not found"

func (s *Service) CreateComment(ctx context.Context, req *pb.CreateCommentReq) (*db.Comment, error) {
	arg := db.CreateCommentParams{
		Comment: req.GetComment(),
		PostID:  req.GetPostId(),
		Author:  req.GetAuthor(),
	}
	comment, err := s.store.CreateComment(ctx, arg)
	if err != nil {
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to create comment: %v", err)
	}
	return &comment, nil
}

func (s *Service) ListCommentsByPost(ctx context.Context, req *pb.GetByIDReq) (*[]db.Comment, error) {
	comments, err := s.store.ListCommentsByPost(ctx, req.GetId())
	if err != nil {
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to list comments: %v", err)
	}
	return &comments, nil
}

func (s *Service) EditComment(ctx context.Context, req *pb.EditCommentReq) (*db.Comment, error) {
	arg := db.EditCommentParams{
		ID:      req.GetId(),
		Comment: req.GetComment(),
	}
	comment, err := s.store.EditComment(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to edit comment: %v", err)
	}
	return &comment, nil
}

func (s *Service) DeleteComment(ctx context.Context, req *pb.GetByIDReq) error {
	err := s.store.DeleteComment(ctx, req.GetId())
	if err != nil {
		if err == pgx.ErrNoRows {
			return grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return grpc_err.ErrorResponse(codes.Internal, "error to delete comment: %v", err)
	}
	return nil
}
