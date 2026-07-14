package posts_service

import (
	"context"

	"github.com/jackc/pgx/v5"
	grpc_err "github.com/koliadertellmi-posts/internal/lib/error/service"
	pb "github.com/koliadertellmi-posts/internal/pb"
	db "github.com/koliadertellmi-posts/internal/store/db/sqlc"
	"google.golang.org/grpc/codes"
)

const notFound = "post not found"

func (s *Service) CreatePost(ctx context.Context, req *pb.CreatePost) (*db.Post, error) {
	arg := db.CreatePostParams{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Author:      req.GetAuthor(),
	}
	post, err := s.store.CreatePost(ctx, arg)
	if err != nil {
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to create post: %v", err)
	}
	return &post, nil
}

func (s *Service) ListPosts(ctx context.Context) (*[]db.Post, error) {
	posts, err := s.store.ListPosts(ctx)
	if err != nil {
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to list posts: %v", err)
	}
	return &posts, nil
}

func (s *Service) GetPostByID(ctx context.Context, req *pb.GetByIDReq) (*db.Post, error) {
	post, err := s.store.GetPostByID(ctx, int32(req.GetId()))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to get post")
	}
	return &post, nil
}

func (s *Service) EditPost(ctx context.Context, req *pb.EditPostReq) (*db.Post, error) {
	arg := db.EditPostParams{
		ID:          int32(req.GetId()),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
	}
	post, err := s.store.EditPost(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to edit post: %v", err)
	}
	return &post, nil
}

func (s *Service) DeletePost(ctx context.Context, req *pb.GetByIDReq) error {
	err := s.store.DeletePost(ctx, int32(req.GetId()))
	if err != nil {
		if err == pgx.ErrNoRows {
			return grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return grpc_err.ErrorResponse(codes.Internal, "error to delete post: %v", err)
	}
	return nil
}
