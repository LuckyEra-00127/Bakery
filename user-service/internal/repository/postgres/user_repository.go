package postgres

import (
	"context"
	"database/sql"

	"github.com/bakeplan/bakeplan-go/user-service/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO users (id, email, password_hash, full_name, role, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, user.ID, user.Email, user.PasswordHash, user.FullName, user.Role, user.CreatedAt)
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, full_name, role, created_at
        FROM users WHERE email = $1
    `, email)
	return scanUser(row)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, full_name, role, created_at
        FROM users WHERE id = $1
    `, id)
	return scanUser(row)
}

func (r *UserRepository) ListClients(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, email, password_hash, full_name, role, created_at
        FROM users WHERE role = 'CLIENT' ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (domain.User, error) {
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role, &user.CreatedAt)
	return user, err
}
