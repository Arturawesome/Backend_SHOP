package web

import (
	pbAuth "BE2_API_Shop/gen/auth"
	rabbitmq "BE2_API_Shop/shop/broker/rebbitmq"
	"BE2_API_Shop/shop/cfg"
	"BE2_API_Shop/shop/internal/data/postgres"
	redisCashe "BE2_API_Shop/shop/internal/data/redis"
	"BE2_API_Shop/shop/web/realtime/sse"
	websockethub "BE2_API_Shop/shop/web/realtime/websocket"

	"BE2_API_Shop/shop/internal/service"

	"context"
	"log"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Application struct {
	services       service.Services
	grpcAuthClient pbAuth.AuthServiceClient
	grpcAuthConn   *grpc.ClientConn
	consumer       *rabbitmq.Consumer
	sseHub         *sse.SSEHub
	wsHub          *websockethub.WebSocketHub
	logger         *slog.Logger
}

func NewApp(cfg *cfg.Config) *Application {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	services, err := Createservices(cfg.Shop.Database.URL, logger, cfg)
	if err != nil {
		log.Fatalf("failed to create services: %v", err)
	}

	conn, err := grpc.NewClient(cfg.Auth.GRPC.ClientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to gRPC: %v", err)
	}
	grpcAuthClient := pbAuth.NewAuthServiceClient(conn)

	sseHub := sse.NewSSEHub()
	wsHub := websockethub.NewWebSocketHub()

	consumer, err := rabbitmq.NewConsumer(*services.Product,
		cfg.RabbitMq.URL, cfg.RabbitMq.Queue, sseHub, wsHub)

	if err != nil {
		log.Fatalf("rebbitmq error: %v", err)
	}
	go func() {
		if err := consumer.Consume(); err != nil {
			logger.Error("rabbitmq consumer stopped", "error", err)
		}
	}()

	return &Application{
		logger:         logger,
		services:       *services,
		grpcAuthClient: grpcAuthClient,
		grpcAuthConn:   conn,
		consumer:       consumer,
		sseHub:         sseHub,
		wsHub:          wsHub,
	}
}

func Createservices(dbUrl string, logger *slog.Logger, cfg *cfg.Config) (*service.Services, error) {
	db, err := postgres.NewDatabase(context.Background(), dbUrl)

	if err != nil {
		logger.Error("database connection failed", "error", err)
		return nil, err
	}

	clientRepo := postgres.NewClientRepository(db.Pool)
	productRepo := postgres.NewProductRepository(db.Pool)
	supplierRepo := postgres.NewSupplierRepository(db.Pool)
	imageRepo := postgres.NewImageRepository(db.Pool)
	reddisClient, err := redisCashe.NewRedisClient(context.Background(), cfg.Redis)
	if err != nil {
		logger.Error("redis connection failed", "error", err)
		return nil, err
	}

	services := service.NewServices(clientRepo, productRepo, supplierRepo, imageRepo, &reddisClient)
	return services, nil
}

func (app *Application) GrpcAuthConnClose() error {
	return app.grpcAuthConn.Close()
}
