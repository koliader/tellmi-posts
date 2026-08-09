package posts_server

import (
	"context"
	"fmt"

	"github.com/koliader/tellmi-posts/internal/lib/config"
	localrabbitmq "github.com/koliader/tellmi-posts/internal/lib/rabbitmq"
	categories_service "github.com/koliader/tellmi-posts/internal/services/categories"
	comments_service "github.com/koliader/tellmi-posts/internal/services/comments"
	posts_service "github.com/koliader/tellmi-posts/internal/services/posts"
	users_service "github.com/koliader/tellmi-posts/internal/services/users"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	redisclient "github.com/koliader/tellmi-posts/internal/store/redis"
	grpcmiddleware "github.com/koliader/tellmi-sdk/middleware"
	pb "github.com/koliader/tellmi-sdk/proto/pb"
	sdkrabbitmq "github.com/koliader/tellmi-sdk/rabbitmq"
	sdktoken "github.com/koliader/tellmi-sdk/token"
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
}

func NewServer(config config.Config, store db.Store, redis *redisclient.Client) (*Server, error) {
	tokenMaker, err := sdktoken.NewJWTMaker(config.TokenKey)
	if err != nil {
		return nil, fmt.Errorf("error to create token maker: %v", err)
	}

	rabbitmqClient, err := sdkrabbitmq.NewClient(config.RbmUrl)
	if err != nil {
		return nil, fmt.Errorf("error to create rabbitmq client: %v", err)
	}

	postsService := posts_service.NewService(store, config, redis)
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
	}

	return &server, nil
}

func (s *Server) CloseRabbitMQ() {
	if s.rabbitmqClient != nil {
		if err := s.rabbitmqClient.Close(); err != nil {
			log.Error().Err(err).Msg("error closing rabbitmq client")
		}
	}
}

func (s *Server) ConsumeUserUpdated(ctx context.Context) error {
	return localrabbitmq.ConsumeUpdateUser(ctx, s.rabbitmqClient, func(ctx context.Context, req sdkrabbitmq.UserUpdated) error {
		log.Info().Msgf("received update user: %s -> %s", req.ID, req.NewUsername)
		user, err := s.users_service.UpdateUser(ctx, &req)
		if err != nil {
			log.Info().Msgf("error to update user: %v", err)
			return err
		}
		log.Info().Msgf("user updated: %v:%s", user.ID, user.Username)
		return nil
	})
}

func (s *Server) ConsumeUserCreated(ctx context.Context) error {
	return localrabbitmq.ConsumeUserCreated(ctx, s.rabbitmqClient, func(ctx context.Context, req sdkrabbitmq.UserCreated) error {
		log.Info().Msgf("received user created: id=%s username=%s", req.ID, req.Username)
		user, err := s.users_service.CreateUser(ctx, &req)
		if err != nil {
			log.Info().Msgf("error to create user: %v", err)
			return err
		}

		log.Info().Msgf("user created: username=%s id=%v", user.Username, user.ID)
		return nil
	})
}
