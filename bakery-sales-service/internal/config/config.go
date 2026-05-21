package config

import "os"

type Config struct {
	Port    string
	DBDSN   string
	NATSURL string
}

func Load() Config {
	return Config{
		Port:    env("SALES_GRPC_PORT", "50052"),
		DBDSN:   env("SALES_DB_DSN", "postgres://sales:password@localhost:5434/sales_db?sslmode=disable"),
		NATSURL: env("NATS_URL", "nats://localhost:4222"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
