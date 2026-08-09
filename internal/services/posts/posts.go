package posts_service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/koliader/tellmi-posts/internal/lib/converter"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	errdb "github.com/koliader/tellmi-sdk/errors/db"
	errsvc "github.com/koliader/tellmi-sdk/errors/service"
	pb "github.com/koliader/tellmi-sdk/proto/pb"
	"github.com/koliader/tellmi-sdk/token"
	"github.com/rs/zerolog/log"
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
	err = s.postsCache.DeleteList(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Internal, "error to delete cache: %v", err)
	}
	return &post, nil
}

func (s *Service) ListPosts(
	ctx context.Context,
	req *pb.PaginationReq,
) (*pb.ListPostsRes, error) {
	limit := req.GetLimit()
	offset := req.GetOffset()

	// 1. Try cache
	cachedPosts, err := s.postsCache.GetList(
		ctx,
		int(limit),
		int(offset),
	)

	if err == nil {
		return &pb.ListPostsRes{
			Posts: cachedPosts,
		}, nil
	}

	// Redis miss is expected.
	// Other Redis errors should be logged.
	if !errors.Is(err, errdb.ErrCacheMiss) {
		log.Warn().
			Err(err).
			Msg("failed to get posts from cache")
	}

	// 2. Get from PostgreSQL
	dbPosts, err := s.store.ListPosts(
		ctx,
		db.ListPostsParams{
			PageLimit:  limit,
			PageOffset: offset,
		},
	)
	if err != nil {
		return nil, errsvc.ErrorResponse(
			codes.Internal,
			"error to list posts: %v",
			err,
		)
	}

	// 3. Convert DB models → protobuf models
	posts := converter.ConvertPostRows(dbPosts)

	// 4. Put into Redis
	if err := s.postsCache.SetList(
		ctx,
		int(limit),
		int(offset),
		posts,
	); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to cache posts")
	}

	// 5. Return protobuf response
	return &pb.ListPostsRes{
		Posts: posts,
	}, nil
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
