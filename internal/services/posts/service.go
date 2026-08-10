package posts_service

import (
	"github.com/koliader/tellmi-posts/internal/lib/config"
	users_service "github.com/koliader/tellmi-posts/internal/services/users"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	redisclient "github.com/koliader/tellmi-posts/internal/store/redis"
)

type Service struct {
	store         db.Store
	config        config.Config
	users_service *users_service.Service
	postsCache    *PostsCache
}

func NewService(store db.Store, config config.Config, redis *redisclient.Client) *Service {
	usersService := users_service.NewService(store, config)
	postsCache := NewPostsCache(redis)
	return &Service{store, config, usersService, postsCache}
}
