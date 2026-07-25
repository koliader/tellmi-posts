package posts_service

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

// TODO ADD FOREIGN KEYS CHECKS
const notFound = "post not found"

func (s *Service) CreatePost(ctx context.Context, req *pb.CreatePostReq, payload *token.Payload) (*db.Post, error) {
	user, err := s.users_service.GetUser(ctx, &payload.Username)
	if err != nil {
		return nil, err
	}
	arg := db.CreatePostParams{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		CategoryID:  req.GetCategoryId(),
		UserID:      user.ID,
	}
	post, err := s.store.CreatePost(ctx, arg)
	if err != nil {
		if db_err.ErrorCode(err) == db_err.ForeignKeyViolation {
			return nil, grpc_err.ErrorResponse(codes.NotFound, "invalid category data")
		}
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to create post: %v", err)
	}
	return &post, nil
}

func (s *Service) ListPosts(ctx context.Context) (*[]db.ListPostsRow, error) {
	posts, err := s.store.ListPosts(ctx)
	if err != nil {
		return nil, grpc_err.ErrorResponse(codes.Internal, "error to list posts: %v", err)
	}
	return &posts, nil
}

func (s *Service) GetPostByID(ctx context.Context, req *pb.GetByIDReq) (*db.Post, error) {
	post, err := s.store.GetPostByID(ctx, req.GetId())
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
		ID:          req.GetId(),
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
	err := s.store.DeletePost(ctx, req.GetId())
	if err != nil {
		if err == pgx.ErrNoRows {
			return grpc_err.ErrorResponse(codes.NotFound, notFound)
		}
		return grpc_err.ErrorResponse(codes.Internal, "error to delete post: %v", err)
	}
	return nil
}
