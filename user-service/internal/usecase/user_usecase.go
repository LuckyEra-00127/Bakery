package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bakeplan/bakeplan-go/user-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	ListClients(ctx context.Context) ([]domain.User, error)
	Delete(ctx context.Context, id string) error
}

type UserUseCase struct {
	repo      UserRepository
	jwtSecret []byte
}

func NewUserUseCase(repo UserRepository, jwtSecret string) *UserUseCase {
	return &UserUseCase{repo: repo, jwtSecret: []byte(jwtSecret)}
}

func (uc *UserUseCase) Register(ctx context.Context, email, password, fullName, role string) (domain.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" || fullName == "" {
		return domain.User{}, "", errors.New("email, password and full_name are required")
	}
	role = strings.ToUpper(strings.TrimSpace(role))
	if role == "" {
		role = "CLIENT"
	}
	if role != "ADMIN" && role != "CLIENT" {
		return domain.User{}, "", errors.New("role must be ADMIN or CLIENT")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, "", err
	}

	user := domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}
	if err := uc.repo.Create(ctx, user); err != nil {
		return domain.User{}, "", err
	}
	token, err := uc.token(user)
	return user, token, err
}

func (uc *UserUseCase) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	user, err := uc.repo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
	if err != nil {
		return domain.User{}, "", errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, "", errors.New("invalid email or password")
	}
	token, err := uc.token(user)
	return user, token, err
}

func (uc *UserUseCase) ValidateToken(ctx context.Context, tokenString string) (domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return uc.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return domain.User{}, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return domain.User{}, errors.New("invalid claims")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return domain.User{}, errors.New("missing user id")
	}
	return uc.repo.GetByID(ctx, sub)
}

func (uc *UserUseCase) GetByID(ctx context.Context, id string) (domain.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UserUseCase) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return uc.repo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
}

func (uc *UserUseCase) ListClients(ctx context.Context) ([]domain.User, error) {
	return uc.repo.ListClients(ctx)
}

func (uc *UserUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *UserUseCase) token(user domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(uc.jwtSecret)
}
