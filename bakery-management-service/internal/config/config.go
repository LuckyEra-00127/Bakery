package config

import "os"

type Config struct {
	Port         string
	DBDSN        string
	NATSURL      string
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

func Load() Config {
	return Config{
		Port:         env("MANAGEMENT_GRPC_PORT", "50053"),
		DBDSN:        env("MANAGEMENT_DB_DSN", "postgres://management:password@localhost:5435/management_db?sslmode=disable"),
		NATSURL:      env("NATS_URL", "nats://localhost:4222"),
		SMTPHost:     env("SMTP_HOST", ""),
		SMTPPort:     env("SMTP_PORT", "587"),
		SMTPUsername: env("SMTP_USERNAME", ""),
		SMTPPassword: env("SMTP_PASSWORD", ""),
		SMTPFrom:     env("SMTP_FROM", "no-reply@bakeplan.local"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
