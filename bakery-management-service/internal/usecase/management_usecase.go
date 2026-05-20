package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	CreateIngredient(ctx context.Context, ingredient domain.Ingredient) error
	ListIngredients(ctx context.Context) ([]domain.Ingredient, error)
	CreateTask(ctx context.Context, task domain.Task) error
	ListTasks(ctx context.Context) ([]domain.Task, error)
	UpdateTask(ctx context.Context, task domain.Task) (domain.Task, error)
	CompleteTask(ctx context.Context, id string) (domain.Task, error)
	DeleteTask(ctx context.Context, id string) error
	ListDailyStatistics(ctx context.Context, date string) ([]domain.DailyStatistic, error)
	LogEmail(ctx context.Context, id, to, subject, body, status string) error
	ListEmailLogs(ctx context.Context) ([]domain.EmailLog, error)
}

type EmailSender interface {
	Enabled() bool
	Send(to, subject, body string) error
}

type ManagementUseCase struct {
	repo   Repository
	sender EmailSender
}

func NewManagementUseCase(repo Repository, sender ...EmailSender) *ManagementUseCase {
	uc := &ManagementUseCase{repo: repo}
	if len(sender) > 0 {
		uc.sender = sender[0]
	}
	return uc
}

func (uc *ManagementUseCase) CreateIngredient(ctx context.Context, name, unit string, cost float64, sections []domain.IngredientSection) (domain.Ingredient, error) {
	name = strings.TrimSpace(name)
	unit = strings.TrimSpace(unit)
	if name == "" || unit == "" || cost < 0 {
		return domain.Ingredient{}, errors.New("name, unit and non-negative cost_per_unit are required")
	}
	ingredient := domain.Ingredient{ID: uuid.NewString(), Name: name, Unit: unit, CostPerUnit: cost, Sections: cleanSections(sections)}
	return ingredient, uc.repo.CreateIngredient(ctx, ingredient)
}

func (uc *ManagementUseCase) ListIngredients(ctx context.Context) ([]domain.Ingredient, error) {
	return uc.repo.ListIngredients(ctx)
}

func (uc *ManagementUseCase) CreateTask(ctx context.Context, title, status, dueDate string) (domain.Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.Task{}, errors.New("title is required")
	}
	if status == "" {
		status = "TODO"
	}
	if !validTaskStatus(status) {
		return domain.Task{}, errors.New("status must be TODO, IN_PROGRESS, DONE or CANCELLED")
	}
	task := domain.Task{ID: uuid.NewString(), Title: title, Status: status, DueDate: dueDate, CreatedAt: time.Now().UTC()}
	return task, uc.repo.CreateTask(ctx, task)
}

func (uc *ManagementUseCase) ListTasks(ctx context.Context) ([]domain.Task, error) {
	return uc.repo.ListTasks(ctx)
}

func (uc *ManagementUseCase) UpdateTask(ctx context.Context, id, title, status, dueDate string) (domain.Task, error) {
	id = strings.TrimSpace(id)
	title = strings.TrimSpace(title)
	status = strings.ToUpper(strings.TrimSpace(status))
	if id == "" || title == "" {
		return domain.Task{}, errors.New("id and title are required")
	}
	if !validTaskStatus(status) {
		return domain.Task{}, errors.New("status must be TODO, IN_PROGRESS, DONE or CANCELLED")
	}
	return uc.repo.UpdateTask(ctx, domain.Task{ID: id, Title: title, Status: status, DueDate: dueDate})
}

func (uc *ManagementUseCase) CompleteTask(ctx context.Context, id string) (domain.Task, error) {
	return uc.repo.CompleteTask(ctx, id)
}

func (uc *ManagementUseCase) DeleteTask(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("task id is required")
	}
	return uc.repo.DeleteTask(ctx, id)
}

func (uc *ManagementUseCase) ListDailyStatistics(ctx context.Context, date string) ([]domain.DailyStatistic, error) {
	return uc.repo.ListDailyStatistics(ctx, date)
}

func (uc *ManagementUseCase) SendEmail(ctx context.Context, to, subject, body string) (id string, success bool, message string, err error) {
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(body)
	if to == "" || subject == "" || body == "" {
		return "", false, "", errors.New("to, subject and body are required")
	}

	id = uuid.NewString()
	status := "DRY_RUN"
	message = "email logged because SMTP is not configured"
	success = true

	if uc.sender != nil && uc.sender.Enabled() {
		if sendErr := uc.sender.Send(to, subject, body); sendErr != nil {
			status = "FAILED"
			message = sendErr.Error()
			success = false
		} else {
			status = "SENT"
			message = "email sent successfully"
			success = true
		}
	}

	if logErr := uc.repo.LogEmail(ctx, id, to, subject, body, status); logErr != nil {
		return "", false, "", logErr
	}
	return id, success, message, nil
}

func (uc *ManagementUseCase) ListEmailLogs(ctx context.Context) ([]domain.EmailLog, error) {
	return uc.repo.ListEmailLogs(ctx)
}

func cleanSections(sections []domain.IngredientSection) []domain.IngredientSection {
	out := make([]domain.IngredientSection, 0, len(sections))
	for _, section := range sections {
		section.Name = strings.TrimSpace(section.Name)
		section.Amount = strings.TrimSpace(section.Amount)
		section.Unit = strings.TrimSpace(section.Unit)
		if section.Name != "" {
			out = append(out, section)
		}
	}
	return out
}

func validTaskStatus(status string) bool {
	switch status {
	case "TODO", "IN_PROGRESS", "DONE", "CANCELLED":
		return true
	default:
		return false
	}
}
