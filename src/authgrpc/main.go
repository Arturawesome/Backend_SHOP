package main

import (
	"BE2_API_Shop/authgrpc/data/postgres"
	"BE2_API_Shop/authgrpc/service"
	pbAuth "BE2_API_Shop/gen/auth"
	"BE2_API_Shop/shop/cfg"

	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := cfg.Load("shop/cfg/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.NewDatabase(context.Background(), cfg.Auth.Database.URL)
	if err != nil {
		fmt.Println("failed to connect to auth database:", err)
		os.Exit(1)
	}
	defer db.Close()

	userRepo := postgres.NewPostgresUserRepository(db.Pool)
	userService := service.NewUserService(userRepo)
	jwtProvider := service.NewJWTProvider(cfg.Auth.JWTSecret)
	authServer := service.NewAuthServer(userService, jwtProvider)

	listener, err := net.Listen("tcp", cfg.Auth.GRPC.ListenAddr)
	if err != nil {
		fmt.Println("failed to listen:", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pbAuth.RegisterAuthServiceServer(grpcServer, authServer)
	if err := grpcServer.Serve(listener); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
