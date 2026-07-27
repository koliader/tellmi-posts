package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	posts_server "github.com/koliader/tellmi-posts/internal/app/grpc/posts"
	"github.com/koliader/tellmi-posts/internal/lib/config"
	"github.com/koliader/tellmi-posts/internal/lib/logger"
	pb "github.com/koliader/tellmi-sdk/proto/pb"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := config.LoadConfig(".")
	// config, err := config.LoadKuberConfig()
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}

	if config.Environment == "dev" || config.Environment == "docker" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	defer connPool.Close()
	store := db.NewStore(connPool)

	go runGrpcServer(config, store)

	server, err := posts_server.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msgf("error creating posts server: %v", err)
	}
	err = server.ConsumeUserUpdated()
	if err != nil {
		log.Fatal().Err(err).Msgf("error starting RabbitMQ update user consumer: %v", err)
	}
	err = server.ConsumeUserCreated()
	if err != nil {
		log.Fatal().Err(err).Msgf("error starting RabbitMQ user created consumer: %v", err)
	}
	select {}
}

func runGrpcServer(config config.Config, store db.Store) {
	postsServer, err := posts_server.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("cannot create posts service: %v", err))
	}
	listener, err := net.Listen("tcp", config.ServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	grpcLogger := grpc.UnaryInterceptor(logger.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterPostsServer(grpcServer, postsServer)
	reflection.Register(grpcServer)

	log.Info().Msgf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start gRPC server")
	}
}
