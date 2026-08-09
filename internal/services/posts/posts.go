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

func (s *Service) CreatePost(
	ctx context.Context,
	req *pb.CreatePostReq,
	payload *token.Payload,
) (*db.Post, error) {
	arg := db.CreatePostParams{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		CategoryID:  req.GetCategoryId(),
		UserID:      payload.ID,
	}

	post, err := s.store.CreatePost(ctx, arg)
	if err != nil {
		if errdb.ErrorCode(err) == errdb.ForeignKeyViolation {
			return nil, errsvc.ErrorResponse(
				codes.NotFound,
				"invalid category or user data: %v",
				err,
			)
		}

		return nil, errsvc.ErrorResponse(
			codes.Internal,
			"error to create post: %v",
			err,
		)
	}

	// Creating a post can change every paginated list.
	if err := s.postsCache.DeleteList(ctx); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to invalidate posts list cache")
	}

	return &post, nil
}

func (s *Service) ListPosts(
	ctx context.Context,
	req *pb.PaginationReq,
) (*pb.ListPostsRes, error) {
	limit := req.GetLimit()
	offset := req.GetOffset()

	// 1. Try Redis cache.
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

	// Cache miss is expected.
	// Other Redis errors should only be logged.
	if !errors.Is(err, errdb.ErrCacheMiss) {
		log.Warn().
			Err(err).
			Msg("failed to get posts from cache")
	}

	// 2. Get posts from PostgreSQL.
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

	// 3. Convert DB models -> protobuf models.
	posts := converter.ConvertPostRows(dbPosts)

	// 4. Store result in Redis.
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

	// 5. Return response.
	return &pb.ListPostsRes{
		Posts: posts,
	}, nil
}

func (s *Service) GetPostByID(
	ctx context.Context,
	req *pb.GetByIDReq,
) (*pb.PostRow, error) {
	id := req.GetId()

	// 1. Try Redis cache.
	cachedPost, err := s.postsCache.GetByID(ctx, id)
	if err == nil {
		return cachedPost, nil
	}

	// Cache miss is expected.
	if !errors.Is(err, errdb.ErrCacheMiss) {
		log.Warn().
			Err(err).
			Msg("failed to get post from cache")
	}

	// 2. Get post from PostgreSQL.
	dbPost, err := s.store.GetPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errsvc.ErrorResponse(
				codes.NotFound,
				notFound,
			)
		}

		return nil, errsvc.ErrorResponse(
			codes.Internal,
			"error to get post: %v",
			err,
		)
	}

	// 3. Convert DB model -> protobuf model.
	post := converter.ConvertGetPostByIDRow(dbPost)

	// 4. Store result in Redis.
	if err := s.postsCache.SetByID(ctx, post); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to cache post")
	}

	// 5. Return response.
	return post, nil
}

func (s *Service) EditPost(
	ctx context.Context,
	req *pb.EditPostReq,
	payload *token.Payload,
) (*db.Post, error) {
	// 1. Get existing post to check ownership.
	post, err := s.store.GetPostByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errsvc.ErrorResponse(
				codes.NotFound,
				notFound,
			)
		}

		return nil, errsvc.ErrorResponse(
			codes.Internal,
			"error to get post: %v",
			err,
		)
	}

	// 2. Check ownership.
	if post.UserID != payload.ID {
		return nil, errsvc.ErrorResponse(
			codes.PermissionDenied,
			"you have no access to change this post",
		)
	}

	// 3. Update PostgreSQL.
	arg := db.EditPostParams{
		ID:          req.GetId(),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		CategoryID:  req.GetCategoryId(),
	}

	updatedPost, err := s.store.EditPost(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errsvc.ErrorResponse(
				codes.NotFound,
				notFound,
			)
		}

		if errdb.ErrorCode(err) == errdb.ForeignKeyViolation {
			return nil, errsvc.ErrorResponse(
				codes.NotFound,
				"invalid category data: %v",
				err,
			)
		}

		return nil, errsvc.ErrorResponse(
			codes.Internal,
			"error to edit post: %v",
			err,
		)
	}

	// 4. Invalidate individual post cache.
	if err := s.postsCache.DeleteByID(ctx, req.GetId()); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to invalidate post cache")
	}

	// 5. Updating a post can change every paginated list.
	if err := s.postsCache.DeleteList(ctx); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to invalidate posts list cache")
	}

	return &updatedPost, nil
}

func (s *Service) DeletePost(
	ctx context.Context,
	req *pb.GetByIDReq,
	payload *token.Payload,
) error {
	// 1. Get post to check ownership.
	post, err := s.store.GetPostByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errsvc.ErrorResponse(
				codes.NotFound,
				notFound,
			)
		}

		return errsvc.ErrorResponse(
			codes.Internal,
			"error to get post: %v",
			err,
		)
	}

	// 2. Check ownership.
	if post.UserID != payload.ID {
		return errsvc.ErrorResponse(
			codes.PermissionDenied,
			"you have no access to delete this post",
		)
	}

	// 3. Delete from PostgreSQL.
	if err := s.store.DeletePost(ctx, req.GetId()); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errsvc.ErrorResponse(
				codes.NotFound,
				notFound,
			)
		}

		return errsvc.ErrorResponse(
			codes.Internal,
			"error to delete post: %v",
			err,
		)
	}

	// 4. Invalidate individual post cache.
	if err := s.postsCache.DeleteByID(ctx, req.GetId()); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to invalidate post cache")
	}

	// 5. Deleting a post can change every paginated list.
	if err := s.postsCache.DeleteList(ctx); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to invalidate posts list cache")
	}

	return nil
}
