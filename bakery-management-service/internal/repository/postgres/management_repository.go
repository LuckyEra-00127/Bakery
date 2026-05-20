package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/domain"
)

type ManagementRepository struct {
	db *sql.DB
}

func NewManagementRepository(db *sql.DB) *ManagementRepository {
	repo := &ManagementRepository{db: db}
	if err := repo.ensureSchema(context.Background()); err != nil {
		log.Printf("management schema check failed: %v", err)
	}
	return repo
}

func (r *ManagementRepository) ensureSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `ALTER TABLE ingredients ADD COLUMN IF NOT EXISTS recipe_sections JSONB NOT NULL DEFAULT '[]'::jsonb`)
	return err
}

func (r *ManagementRepository) CreateIngredient(ctx context.Context, ingredient domain.Ingredient) error {
	sections, err := json.Marshal(ingredient.Sections)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO ingredients (id, name, unit, cost_per_unit, recipe_sections) VALUES ($1, $2, $3, $4, $5::jsonb)`, ingredient.ID, ingredient.Name, ingredient.Unit, ingredient.CostPerUnit, string(sections))
	return err
}

func (r *ManagementRepository) ListIngredients(ctx context.Context) ([]domain.Ingredient, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, unit, cost_per_unit, COALESCE(recipe_sections::text, '[]') FROM ingredients ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ingredients []domain.Ingredient
	for rows.Next() {
		var i domain.Ingredient
		var sectionsJSON string
		if err := rows.Scan(&i.ID, &i.Name, &i.Unit, &i.CostPerUnit, &sectionsJSON); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(sectionsJSON), &i.Sections)
		ingredients = append(ingredients, i)
	}
	return ingredients, rows.Err()
}

func (r *ManagementRepository) CreateTask(ctx context.Context, task domain.Task) error {
	var due any
	if task.DueDate != "" {
		due = task.DueDate
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO tasks (id, title, status, due_date, created_at) VALUES ($1, $2, $3, $4, $5)`, task.ID, task.Title, task.Status, due, task.CreatedAt)
	return err
}

func (r *ManagementRepository) ListTasks(ctx context.Context) ([]domain.Task, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, title, status, COALESCE(due_date::text, ''), created_at FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.DueDate, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *ManagementRepository) UpdateTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	var due any
	if task.DueDate != "" {
		due = task.DueDate
	}
	row := r.db.QueryRowContext(ctx, `UPDATE tasks SET title = $2, status = $3, due_date = $4 WHERE id = $1 RETURNING id, title, status, COALESCE(due_date::text, ''), created_at`, task.ID, task.Title, task.Status, due)
	var updated domain.Task
	err := row.Scan(&updated.ID, &updated.Title, &updated.Status, &updated.DueDate, &updated.CreatedAt)
	return updated, err
}

func (r *ManagementRepository) CompleteTask(ctx context.Context, id string) (domain.Task, error) {
	row := r.db.QueryRowContext(ctx, `UPDATE tasks SET status = 'DONE' WHERE id = $1 RETURNING id, title, status, COALESCE(due_date::text, ''), created_at`, id)
	var t domain.Task
	err := row.Scan(&t.ID, &t.Title, &t.Status, &t.DueDate, &t.CreatedAt)
	return t, err
}

func (r *ManagementRepository) DeleteTask(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("task not found")
	}
	return nil
}

func (r *ManagementRepository) ListDailyStatistics(ctx context.Context, date string) ([]domain.DailyStatistic, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT stat_date::text, product_id, COALESCE(product_name,''), baked, delivered, sold, returned, left_qty, revenue
        FROM statistics_daily
        WHERE ($1 = '' OR stat_date = $1::date)
        ORDER BY stat_date DESC
    `, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []domain.DailyStatistic
	for rows.Next() {
		var s domain.DailyStatistic
		if err := rows.Scan(&s.Date, &s.ProductID, &s.Product, &s.Baked, &s.Delivered, &s.Sold, &s.Returned, &s.Left, &s.Revenue); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *ManagementRepository) LogEmail(ctx context.Context, id, to, subject, body, status string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO email_logs (id, recipient, subject, body, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)`, id, to, subject, body, status, time.Now().UTC())
	return err
}

func (r *ManagementRepository) ListEmailLogs(ctx context.Context) ([]domain.EmailLog, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, recipient, subject, body, status, created_at FROM email_logs ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []domain.EmailLog
	for rows.Next() {
		var logEntry domain.EmailLog
		if err := rows.Scan(&logEntry.ID, &logEntry.Recipient, &logEntry.Subject, &logEntry.Body, &logEntry.Status, &logEntry.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}
	return logs, rows.Err()
}
