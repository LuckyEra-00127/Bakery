package config

import "os"

type Config struct {
	Port      string
	DBDSN     string
	JWTSecret string
}

func Load() Config {
	return Config{
		Port:      env("USER_GRPC_PORT", "50051"),
		DBDSN:     env("USER_DB_DSN", "postgres://user:password@localhost:5433/user_db?sslmode=disable"),
		JWTSecret: env("JWT_SECRET", "change-me-super-secret"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
