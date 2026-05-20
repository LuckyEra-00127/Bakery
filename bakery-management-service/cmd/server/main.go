package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/config"
	grpcdelivery "github.com/bakeplan/bakeplan-go/bakery-management-service/internal/delivery/grpc"
	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/email"
	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/messaging"
	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/repository/postgres"
	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/usecase"
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

	listener := messaging.NewListener(cfg.NATSURL)
	listener.Start()
	defer listener.Close()

	repo := postgres.NewManagementRepository(db)
	mailer := email.NewSMTPSender(email.SMTPConfig{Host: cfg.SMTPHost, Port: cfg.SMTPPort, Username: cfg.SMTPUsername, Password: cfg.SMTPPassword, From: cfg.SMTPFrom})
	uc := usecase.NewManagementUseCase(repo, mailer)

	tcp, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	grpcdelivery.Register(server, grpcdelivery.NewServer(uc))

	log.Printf("bakery-management-service listening on :%s", cfg.Port)
	if err := server.Serve(tcp); err != nil {
		log.Fatal(err)
	}
}
