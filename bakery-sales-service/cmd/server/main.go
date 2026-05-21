package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/config"
	grpcdelivery "github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/delivery/grpc"
	"github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/messaging"
	"github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/repository/postgres"
	"github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/usecase"
	"github.com/bakeplan/bakeplan-go/shared/grpcjson"
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

	pub := messaging.NewPublisher(cfg.NATSURL)
	defer pub.Close()

	repo := postgres.NewSalesRepository(db)
	uc := usecase.NewSalesUseCase(repo, pub)

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	grpcdelivery.Register(server, grpcdelivery.NewServer(uc))

	log.Printf("bakery-sales-service listening on :%s", cfg.Port)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
