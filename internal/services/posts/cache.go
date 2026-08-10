package posts_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	redisclient "github.com/koliader/tellmi-posts/internal/store/redis"
	errdb "github.com/koliader/tellmi-sdk/errors/db"
	"github.com/koliader/tellmi-sdk/proto/pb"
	"github.com/redis/go-redis/v9"
)

// The feed is a single, persistent Redis list holding the newest posts in
// order (index 0 = newest). It has no TTL: entries only change through
// explicit operations below, so the key stays visible in Redis and never
// requires a cache fill on a hot read.
const feedDepth = 300

const feedKey = "posts:feed"

type PostsCache struct {
	redis *redisclient.Client
}

func NewPostsCache(redis *redisclient.Client) *PostsCache {
	return &PostsCache{
		redis: redis,
	}
}

// Keys

func postByIDKey(id int64) string {
	return fmt.Sprintf("posts:byid:%d", id)
}

// Feed

// GetFeed returns the posts in the feed window [offset, offset+limit).
// found is false when the feed is not yet populated to feedDepth or the
// requested window falls outside it, in which case the caller should refill
// the feed from PostgreSQL.
func (c *PostsCache) GetFeed(
	ctx context.Context,
	limit int,
	offset int,
) ([]*pb.PostRow, bool, error) {
	n, err := c.redis.LLen(ctx, feedKey)
	if err != nil {
		return nil, false, fmt.Errorf("get feed length: %w", err)
	}

	// A partially populated feed (fewer than feedDepth entries) may be missing
	// posts that exist in PostgreSQL, so treat it as a miss and refill.
	if n < feedDepth || int64(offset+limit) > n {
		return nil, false, nil
	}

	elements, err := c.redis.LRange(ctx, feedKey, int64(offset), int64(offset+limit-1))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get feed window: %w", err)
	}

	posts := make([]*pb.PostRow, 0, len(elements))
	for _, el := range elements {
		var post pb.PostRow
		if err := json.Unmarshal([]byte(el), &post); err != nil {
			return nil, false, fmt.Errorf("unmarshal feed post: %w", err)
		}
		posts = append(posts, &post)
	}

	return posts, true, nil
}

// SetFeed replaces the feed with posts ordered newest-first.
func (c *PostsCache) SetFeed(
	ctx context.Context,
	posts []*pb.PostRow,
) error {
	elements := make([]string, 0, len(posts))
	for _, post := range posts {
		data, err := json.Marshal(post)
		if err != nil {
			return fmt.Errorf("marshal feed post: %w", err)
		}
		elements = append(elements, string(data))
	}

	pipe := c.redis.Pipeline(ctx)
	pipe.Del(ctx, feedKey)
	if len(elements) > 0 {
		values := make([]interface{}, len(elements))
		for i, el := range elements {
			values[i] = el
		}
		pipe.RPush(ctx, feedKey, values...)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("set feed: %w", err)
	}

	return nil
}

// PrependPost atomically pushes a newly created post to the head of the feed
// and trims it back to feedDepth, so it is immediately visible in the list
// with no delay and no cache invalidation.
func (c *PostsCache) PrependPost(
	ctx context.Context,
	post *pb.PostRow,
) error {
	data, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("marshal feed post: %w", err)
	}

	if err := c.redis.LPush(ctx, feedKey, string(data)); err != nil {
		return fmt.Errorf("prepend post to feed: %w", err)
	}

	if err := c.redis.LTrim(ctx, feedKey, 0, feedDepth-1); err != nil {
		return fmt.Errorf("trim feed: %w", err)
	}

	return nil
}

// DeleteFeed drops the whole feed. Used on edit/delete where the existing
// windowed entries are no longer trustworthy; the next read refills it.
func (c *PostsCache) DeleteFeed(ctx context.Context) error {
	if err := c.redis.Delete(ctx, feedKey); err != nil {
		return fmt.Errorf("delete feed: %w", err)
	}

	return nil
}

// By ID

func (c *PostsCache) GetByID(
	ctx context.Context,
	id int64,
) (*pb.PostRow, error) {
	key := postByIDKey(id)

	data, err := c.redis.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errdb.ErrCacheMiss
		}

		return nil, fmt.Errorf("get post from cache: %w", err)
	}

	var post pb.PostRow

	if err := json.Unmarshal([]byte(data), &post); err != nil {
		return nil, fmt.Errorf("unmarshal cached post: %w", err)
	}

	return &post, nil
}

func (c *PostsCache) SetByID(
	ctx context.Context,
	post *pb.PostRow,
) error {
	data, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("marshal post: %w", err)
	}

	key := postByIDKey(post.GetId())

	if err := c.redis.Set(ctx, key, string(data)); err != nil {
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
