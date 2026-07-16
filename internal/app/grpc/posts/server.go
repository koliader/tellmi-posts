package posts_server

import (
	"fmt"

	"github.com/koliader/tellmi-posts/internal/lib/config"
	"github.com/koliader/tellmi-posts/internal/lib/middleware"
	"github.com/koliader/tellmi-posts/internal/lib/token"
	pb "github.com/koliader/tellmi-posts/internal/pb"
	categories_service "github.com/koliader/tellmi-posts/internal/services/categories"
	comments_service "github.com/koliader/tellmi-posts/internal/services/comments"
	posts_service "github.com/koliader/tellmi-posts/internal/services/posts"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
)

type Server struct {
	pb.UnimplementedPostsServer
	posts_service      posts_service.Service
	categories_service categories_service.Service
	comments_service   comments_service.Service
	middleware         middleware.Middleware
}

func NewServer(config config.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenKey)
	if err != nil {
		return nil, fmt.Errorf("error to create token maker: %v", err)
	}

	postsService := posts_service.NewService(store, config)
	categoriesService := categories_service.NewServer(config, store)
	commentsService := comments_service.NewService(store, config)
	middleware := middleware.NewMiddleware(tokenMaker)

	server := Server{
		posts_service:      *postsService,
		categories_service: *categoriesService,
		comments_service:   *commentsService,
		middleware:          *middleware,
	}

	return &server, nil
}
