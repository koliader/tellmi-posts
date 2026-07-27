package posts_server

import (
	"context"
	"fmt"

	"github.com/koliader/tellmi-posts/internal/lib/config"
	grpcmiddleware "github.com/koliader/tellmi-sdk/middleware"
	sdkrabbitmq "github.com/koliader/tellmi-sdk/rabbitmq"
	"github.com/koliader/tellmi-sdk/token"
	pb "github.com/koliader/tellmi-sdk/proto/pb"
	categories_service "github.com/koliader/tellmi-posts/internal/services/categories"
	comments_service "github.com/koliader/tellmi-posts/internal/services/comments"
	localrabbitmq "github.com/koliader/tellmi-posts/internal/lib/rabbitmq"
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
	middleware         *grpcmiddleware.GrpcMiddleware
	rabbitmqClient     *sdkrabbitmq.Client
	users_service      *users_service.Service
	token_maker        token.Maker
}

func NewServer(config config.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenKey)
	if err != nil {
		return nil, fmt.Errorf("error to create token maker: %v", err)
	}

	rabbitmqClient, err := sdkrabbitmq.NewClient(config.RbmUrl)
	if err != nil {
		return nil, fmt.Errorf("error to create rabbitmq client: %v", err)
	}

	postsService := posts_service.NewService(store, config)
	categoriesService := categories_service.NewServer(config, store)
	commentsService := comments_service.NewService(store, config)
	usersService := users_service.NewService(store, config)
	mw := grpcmiddleware.NewMiddleware(tokenMaker)

	server := Server{
		posts_service:      *postsService,
		categories_service: *categoriesService,
		comments_service:   *commentsService,
		middleware:         mw,
		rabbitmqClient:     rabbitmqClient,
		users_service:      usersService,
		token_maker:        tokenMaker,
	}

	return &server, nil
}

func (s *Server) ConsumeUserUpdated() error {
	return localrabbitmq.ConsumeUpdateUser(s.rabbitmqClient, func(req sdkrabbitmq.UserUpdated) error {
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
	return localrabbitmq.ConsumeUserCreated(s.rabbitmqClient, func(req sdkrabbitmq.UserCreated) error {
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
