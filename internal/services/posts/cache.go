package posts_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redisclient "github.com/koliader/tellmi-posts/internal/store/redis"
	errdb "github.com/koliader/tellmi-sdk/errors/db"
	"github.com/koliader/tellmi-sdk/proto/pb"
	"github.com/redis/go-redis/v9"
)

const (
	postsListTTL = 5 * time.Second
	postByIDTTL  = 5 * time.Second
)

type PostsCache struct {
	redis *redisclient.Client
}

func NewPostsCache(redis *redisclient.Client) *PostsCache {
	return &PostsCache{
		redis: redis,
	}
}

// Keys

func postsListKey(limit, offset int) string {
	return fmt.Sprintf("posts:list:%d:%d", limit, offset)
}

func postByIDKey(id int64) string {
	return fmt.Sprintf("posts:byid:%d", id)
}

// List

func (c *PostsCache) GetList(
	ctx context.Context,
	limit int,
	offset int,
) ([]pb.Post, error) {
	key := postsListKey(limit, offset)

	data, err := c.redis.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errdb.ErrCacheMiss
		}

		return nil, fmt.Errorf("get posts from cache: %w", err)
	}

	var posts []pb.Post

	if err := json.Unmarshal([]byte(data), &posts); err != nil {
		return nil, fmt.Errorf("unmarshal cached posts: %w", err)
	}

	return posts, nil
}

func (c *PostsCache) SetList(
	ctx context.Context,
	limit int,
	offset int,
	posts []pb.Post,
) error {
	data, err := json.Marshal(posts)
	if err != nil {
		return fmt.Errorf("marshal posts: %w", err)
	}

	key := postsListKey(limit, offset)

	if err := c.redis.Set(
		ctx,
		key,
		string(data),
		postsListTTL,
	); err != nil {
		return fmt.Errorf("set posts cache: %w", err)
	}

	return nil
}

// By ID
func (c *PostsCache) GetByID(
	ctx context.Context,
	id int64,
) (*pb.Post, error) {
	key := postByIDKey(id)

	data, err := c.redis.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errdb.ErrCacheMiss
		}

		return nil, fmt.Errorf("get post from cache: %w", err)
	}

	var post pb.Post

	if err := json.Unmarshal([]byte(data), &post); err != nil {
		return nil, fmt.Errorf("unmarshal cached post: %w", err)
	}

	return &post, nil
}

func (c *PostsCache) SetByID(
	ctx context.Context,
	post *pb.Post,
) error {
	data, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("marshal post: %w", err)
	}

	key := postByIDKey(post.GetId())

	if err := c.redis.Set(
		ctx,
		key,
		string(data),
		postByIDTTL,
	); err != nil {
		return fmt.Errorf("set post cache: %w", err)
	}

	return nil
}

func (c *PostsCache) DeleteByID(
	ctx context.Context,
	id int64,
) error {
	key := postByIDKey(id)

	if err := c.redis.Delete(ctx, key); err != nil {
		return fmt.Errorf("delete post cache: %w", err)
	}

	return nil
}
