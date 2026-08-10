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

	// Push the new post to the head of the feed immediately (atomic LPUSH +
	// LTRIM). This makes it visible in the list with no delay and no
	// invalidation of the cached feed.
	if row, err := s.store.GetPostByID(ctx, post.ID); err == nil {
		postRow := converter.ConvertGetPostByIDRow(row)
		if err := s.postsCache.SetByID(ctx, postRow); err != nil {
			log.Warn().Err(err).Msg("failed to cache created post")
		}
		if err := s.postsCache.PrependPost(ctx, postRow); err != nil {
			log.Warn().Err(err).Msg("failed to prepend created post to feed")
		}
	} else {
		log.Warn().Err(err).Int64("post_id", post.ID).Msg("failed to load created post for cache")
	}

	return &post, nil
}

func (s *Service) ListPosts(
	ctx context.Context,
	req *pb.PaginationReq,
) (*pb.ListPostsRes, error) {
	limit := req.GetLimit()
	offset := req.GetOffset()

	// Windows inside the cached feed are served straight from Redis. Deep
	// pagination beyond the feed depth goes directly to PostgreSQL.
	if int(offset)+int(limit) <= feedDepth {
		cachedPosts, found, err := s.postsCache.GetFeed(ctx, int(limit), int(offset))
		if err != nil {
			log.Warn().Err(err).Msg("failed to get posts feed from cache")
		}
		if found {
			return &pb.ListPostsRes{
				Posts: cachedPosts,
			}, nil
		}
	}

	// 2. Get posts from PostgreSQL (refill the whole feed depth on a miss).
	dbPosts, err := s.store.ListPosts(
		ctx,
		db.ListPostsParams{
			PageLimit:  int32(feedDepth),
			PageOffset: 0,
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

	// 4. Refill the feed and serve the requested window from it.
	if int(offset)+int(limit) <= feedDepth {
		if err := s.postsCache.SetFeed(ctx, posts); err != nil {
			log.Warn().Err(err).Msg("failed to cache posts feed")
		}

		start := min(offset, int32(len(posts)))
		end := min(offset+limit, int32(len(posts)))
		posts = posts[start:end]
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

	// 5. Updating a post can change the cached feed.
	if err := s.postsCache.DeleteFeed(ctx); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to invalidate posts feed cache")
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

	// 5. Deleting a post shifts the cached feed.
	if err := s.postsCache.DeleteFeed(ctx); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to invalidate posts feed cache")
	}

	return nil
}
