package domain

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	Role         string
	CreatedAt    time.Time
}
