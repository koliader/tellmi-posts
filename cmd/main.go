package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	posts_server "github.com/koliader/tellmi-posts/internal/app/grpc/posts"
	"github.com/koliader/tellmi-posts/internal/lib/config"
	"github.com/koliader/tellmi-posts/internal/lib/logger"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	redisclient "github.com/koliader/tellmi-posts/internal/store/redis"
	"github.com/koliader/tellmi-sdk/health"
	"github.com/koliader/tellmi-sdk/otel"
	pb "github.com/koliader/tellmi-sdk/proto/pb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}

	if config.Environment == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(otel.NewZerologHook())
	} else {
		log.Logger = log.Logger.Hook(otel.NewZerologHook())
	}
	zerolog.DefaultContextLogger = &log.Logger

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otelSDK, err := otel.Init(ctx, otel.Config{ServiceName: "posts-service"})
	if err != nil {
		log.Fatal().Err(err).Msg("cannot init otel")
	}
	defer func() {
		if err := otelSDK.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("otel shutdown")
		}
	}()

	pgxConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot parse db config")
	}
	pgxConfig.ConnConfig.Tracer = otelpgx.NewTracer()
	connPool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	defer connPool.Close()
	store := db.NewStore(connPool)

	redisClient := redisclient.New(ctx, redisclient.Config{Addr: config.RedisUrl, Password: "", DB: config.RedisDBNumber})
	defer redisClient.Close()

	if config.HealthAddress != "" {
		healthServer := health.NewServer(config.HealthAddress,
			func(ctx context.Context) error {
				return connPool.Ping(ctx)
			},
		).WithHandler("/metrics", otel.MetricsHandler())
		go func() {
			log.Info().Msgf("start health server at %s", config.HealthAddress)
			if err := healthServer.Start(); err != nil {
				log.Error().Err(err).Msg("health server stopped")
			}
		}()
	}

	server, err := posts_server.NewServer(config, store, redisClient)
	if err != nil {
		log.Fatal().Err(err).Msgf("error creating posts server: %v", err)
	}

	grpcServer := runGrpcServer(server, config, otelSDK.TracerProvider, otelSDK.MeterProvider)

	go func() {
		if err := server.ConsumeUserUpdated(ctx); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("RabbitMQ update user consumer stopped")
		}
	}()
	go func() {
		if err := server.ConsumeUserCreated(ctx); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("RabbitMQ user created consumer stopped")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down posts service")
	server.CloseRabbitMQ()
	grpcServer.GracefulStop()
}

func runGrpcServer(postsServer *posts_server.Server, config config.Config, tracerProvider trace.TracerProvider, meterProvider metric.MeterProvider) *grpc.Server {
	listener, err := net.Listen("tcp", config.ServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	grpcLogger := grpc.UnaryInterceptor(logger.GrpcLogger)
	grpcServer := grpc.NewServer(
		grpcLogger,
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(tracerProvider),
			otelgrpc.WithMeterProvider(meterProvider),
		)),
	)
	pb.RegisterPostsServer(grpcServer, postsServer)
	reflection.Register(grpcServer)

	log.Info().Msgf("start gRPC server at %s", listener.Addr().String())
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal().Err(err).Msg("cannot start gRPC server")
		}
	}()

	return grpcServer
}
