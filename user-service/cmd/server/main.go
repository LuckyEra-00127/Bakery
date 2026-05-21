package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/bakeplan/bakeplan-go/shared/grpcjson"
	"github.com/bakeplan/bakeplan-go/user-service/internal/config"
	grpcdelivery "github.com/bakeplan/bakeplan-go/user-service/internal/delivery/grpc"
	"github.com/bakeplan/bakeplan-go/user-service/internal/repository/postgres"
	"github.com/bakeplan/bakeplan-go/user-service/internal/usecase"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	grpcjson.Register()
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	repo := postgres.NewUserRepository(db)
	uc := usecase.NewUserUseCase(repo, cfg.JWTSecret)

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	grpcdelivery.Register(server, grpcdelivery.NewServer(uc))

	log.Printf("user-service listening on :%s", cfg.Port)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
