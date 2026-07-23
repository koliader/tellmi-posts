package posts_server

import (
	"context"
	"fmt"

	"github.com/koliader/tellmi-posts/internal/lib/config"
	"github.com/koliader/tellmi-posts/internal/lib/middleware"
	"github.com/koliader/tellmi-posts/internal/lib/rabbitmq"
	"github.com/koliader/tellmi-posts/internal/lib/token"
	pb "github.com/koliader/tellmi-posts/internal/pb"
	categories_service "github.com/koliader/tellmi-posts/internal/services/categories"
	comments_service "github.com/koliader/tellmi-posts/internal/services/comments"
	posts_service "github.com/koliader/tellmi-posts/internal/services/posts"
	users_service "github.com/koliader/tellmi-posts/internal/services/users"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	"github.com/rs/zerolog/log"
)

type Server struct {
	pb.UnimplementedPostsServer
	posts_service      posts_service.Service
	categories_service categories_service.Service
	comments_service   comments_service.Service
	middleware         middleware.Middleware
	rabbitmqClient     *rabbitmq.Client
	users_service      *users_service.Service
	token_maker        token.Maker
}

func NewServer(config config.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenKey)
	if err != nil {
		return nil, fmt.Errorf("error to create token maker: %v", err)
	}

	rabbitmqClient, err := rabbitmq.NewRabbitmqClient(config)
	if err != nil {
		return nil, fmt.Errorf("error to create rabbitmq client: %v", err)
	}

	postsService := posts_service.NewService(store, config)
	categoriesService := categories_service.NewServer(config, store)
	commentsService := comments_service.NewService(store, config)
	usersService := users_service.NewService(config, store)
	middleware := middleware.NewMiddleware(tokenMaker)

	server := Server{
		posts_service:      *postsService,
		categories_service: *categoriesService,
		comments_service:   *commentsService,
		middleware:         *middleware,
		rabbitmqClient:     rabbitmqClient,
		users_service:      usersService,
		token_maker:        tokenMaker,
	}

	return &server, nil
}

func (s *Server) ConsumeUserUpdated() error {
	return s.rabbitmqClient.ConsumeUpdateUser(func(req rabbitmq.UserUpdated) error {
		log.Info().Msgf("received update user username: %v -> %s", req.Username, req.NewUsername)
		user, err := s.users_service.UpdateUser(context.Background(), &req)
		if err != nil {
			log.Info().Msgf("error to update user: %v", err)
			return err
		}
		log.Info().Msgf("user updated: %v:%s", user.ID, user.Username)
		return nil
	})
}

func (s *Server) ConsumeUserCreated() error {
	return s.rabbitmqClient.ConsumeUserCreated(func(req rabbitmq.UserCreated) error {
		log.Info().Msgf("received user created: username=%s", req.Username)
		// TODO: implement actual logic for new user creation in posts

		user, err := s.users_service.CreateUser(context.Background(), &req)
		if err != nil {
			log.Info().Msgf("error to create user: %v", err)
			return err
		}

		log.Info().Msgf("user created: username=%s id=%v", user.Username, user.ID)
		return nil
	})
}
