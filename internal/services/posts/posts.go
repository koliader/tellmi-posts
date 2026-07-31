package posts_service

import (
	"context"

	"github.com/jackc/pgx/v5"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	errdb "github.com/koliader/tellmi-sdk/errors/db"
	errsvc "github.com/koliader/tellmi-sdk/errors/service"
	pb "github.com/koliader/tellmi-sdk/proto/pb"
	"github.com/koliader/tellmi-sdk/token"
	"google.golang.org/grpc/codes"
)

const notFound = "post not found"

func (s *Service) CreatePost(ctx context.Context, req *pb.CreatePostReq, payload *token.Payload) (*db.Post, error) {
	arg := db.CreatePostParams{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		CategoryID:  req.GetCategoryId(),
		UserID:      payload.ID,
	}
	post, err := s.store.CreatePost(ctx, arg)
	if err != nil {
		if errdb.ErrorCode(err) == errdb.ForeignKeyViolation {
			return nil, errsvc.ErrorResponse(codes.NotFound, "invalid category or user data: %v", err)
		}
		return nil, errsvc.ErrorResponse(codes.Internal, "error to create post: %v", err)
	}
	return &post, nil
}

func (s *Service) ListPosts(ctx context.Context) (*[]db.ListPostsRow, error) {
	posts, err := s.store.ListPosts(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Internal, "error to list posts: %v", err)
	}
	return &posts, nil
}

func (s *Service) GetPostByID(ctx context.Context, req *pb.GetByIDReq) (*db.GetPostByIDRow, error) {
	post, err := s.store.GetPostByID(ctx, req.GetId())
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errsvc.ErrorResponse(codes.NotFound, notFound)
		}
		return nil, errsvc.ErrorResponse(codes.Internal, "error to get post")
	}
	return &post, nil
}

func (s *Service) EditPost(ctx context.Context, req *pb.EditPostReq, payload *token.Payload) (*db.Post, error) {
	post, err := s.store.GetPostByID(ctx, req.Id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errsvc.ErrorResponse(codes.NotFound, notFound)
		}
		return nil, errsvc.ErrorResponse(codes.Internal, "error to get post: %v", err)
	}
	if post.UserID != payload.ID {
		return nil, errsvc.ErrorResponse(codes.PermissionDenied, "you have no access to change this post")
	}
	arg := db.EditPostParams{
		ID:          req.GetId(),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		CategoryID:  req.GetCategoryId(),
	}
	updatedtPost, err := s.store.EditPost(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errsvc.ErrorResponse(codes.NotFound, notFound)
		}
		if errdb.ErrorCode(err) == errdb.ForeignKeyViolation {
			return nil, errsvc.ErrorResponse(codes.NotFound, "invalid category data: %v", err)
		}
		return nil, errsvc.ErrorResponse(codes.Internal, "error to edit post: %v", err)
	}
	return &updatedtPost, nil
}

func (s *Service) DeletePost(ctx context.Context, req *pb.GetByIDReq, payload *token.Payload) error {
	post, err := s.GetPostByID(ctx, req)
	if err != nil {
		return err
	}
	if post.UserID != payload.ID {
		return errsvc.ErrorResponse(codes.PermissionDenied, "you have no access to delete this post")
	}
	err = s.store.DeletePost(ctx, req.GetId())
	if err != nil {
		if err == pgx.ErrNoRows {
			return errsvc.ErrorResponse(codes.NotFound, notFound)
		}
		return errsvc.ErrorResponse(codes.Internal, "error to delete post: %v", err)
	}
	return nil
}
