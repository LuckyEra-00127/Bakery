package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/domain"
)

type fakeManagementRepo struct {
	ingredients []domain.Ingredient
	tasks       map[string]domain.Task
	emailLogs   []domain.EmailLog
}

func (r *fakeManagementRepo) CreateIngredient(_ context.Context, ingredient domain.Ingredient) error {
	r.ingredients = append(r.ingredients, ingredient)
	return nil
}

func (r *fakeManagementRepo) ListIngredients(_ context.Context) ([]domain.Ingredient, error) {
	return r.ingredients, nil
}

func (r *fakeManagementRepo) CreateTask(_ context.Context, task domain.Task) error {
	if r.tasks == nil {
		r.tasks = map[string]domain.Task{}
	}
	r.tasks[task.ID] = task
	return nil
}

func (r *fakeManagementRepo) ListTasks(_ context.Context) ([]domain.Task, error) {
	var tasks []domain.Task
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *fakeManagementRepo) UpdateTask(_ context.Context, task domain.Task) (domain.Task, error) {
	if r.tasks == nil || r.tasks[task.ID].ID == "" {
		return domain.Task{}, errors.New("task not found")
	}
	existing := r.tasks[task.ID]
	existing.Title = task.Title
	existing.Status = task.Status
	existing.DueDate = task.DueDate
	r.tasks[task.ID] = existing
	return existing, nil
}

func (r *fakeManagementRepo) CompleteTask(_ context.Context, id string) (domain.Task, error) {
	if r.tasks == nil || r.tasks[id].ID == "" {
		return domain.Task{}, errors.New("task not found")
	}
	task := r.tasks[id]
	task.Status = "DONE"
	r.tasks[id] = task
	return task, nil
}

func (r *fakeManagementRepo) DeleteTask(_ context.Context, id string) error {
	if r.tasks == nil || r.tasks[id].ID == "" {
		return errors.New("task not found")
	}
	delete(r.tasks, id)
	return nil
}

func (r *fakeManagementRepo) ListDailyStatistics(_ context.Context, _ string) ([]domain.DailyStatistic, error) {
	return nil, nil
}

func (r *fakeManagementRepo) LogEmail(_ context.Context, id, to, subject, body, status string) error {
	r.emailLogs = append(r.emailLogs, domain.EmailLog{ID: id, Recipient: to, Subject: subject, Body: body, Status: status, CreatedAt: time.Now().UTC()})
	return nil
}

func (r *fakeManagementRepo) ListEmailLogs(_ context.Context) ([]domain.EmailLog, error) {
	return r.emailLogs, nil
}

type fakeEmailSender struct {
	enabled bool
	sent    bool
	err     error
}

func (s *fakeEmailSender) Enabled() bool { return s.enabled }
func (s *fakeEmailSender) Send(_, _, _ string) error {
	s.sent = true
	return s.err
}

func TestCreateIngredientCleansSections(t *testing.T) {
	repo := &fakeManagementRepo{}
	uc := NewManagementUseCase(repo)
	ingredient, err := uc.CreateIngredient(context.Background(), " Bread ", " 1 kg ", 200, []domain.IngredientSection{
		{Name: " Salt ", Amount: " 10 ", Unit: " gr "},
		{Name: " ", Amount: "100", Unit: "gr"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ingredient.Name != "Bread" || len(ingredient.Sections) != 1 || ingredient.Sections[0].Name != "Salt" {
		t.Fatalf("ingredient was not cleaned correctly: %+v", ingredient)
	}
}

func TestTaskStatusValidation(t *testing.T) {
	repo := &fakeManagementRepo{}
	uc := NewManagementUseCase(repo)
	task, err := uc.CreateTask(context.Background(), "Clean oven", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != "TODO" {
		t.Fatalf("expected default TODO status, got %s", task.Status)
	}
	if _, err := uc.UpdateTask(context.Background(), task.ID, task.Title, "bad", ""); err == nil {
		t.Fatal("expected invalid task status error")
	}
}

func TestSendEmailDryRunAndSMTP(t *testing.T) {
	repo := &fakeManagementRepo{}
	dryRunUC := NewManagementUseCase(repo)
	_, success, message, err := dryRunUC.SendEmail(context.Background(), "client@example.com", "Order", "Ready")
	if err != nil || !success || message == "" {
		t.Fatalf("expected dry-run success, success=%v message=%q err=%v", success, message, err)
	}
	if repo.emailLogs[0].Status != "DRY_RUN" {
		t.Fatalf("expected DRY_RUN log, got %s", repo.emailLogs[0].Status)
	}

	sender := &fakeEmailSender{enabled: true}
	repo = &fakeManagementRepo{}
	smtpUC := NewManagementUseCase(repo, sender)
	_, success, _, err = smtpUC.SendEmail(context.Background(), "client@example.com", "Order", "Ready")
	if err != nil || !success || !sender.sent {
		t.Fatalf("expected SMTP send success, success=%v sent=%v err=%v", success, sender.sent, err)
	}
	if repo.emailLogs[0].Status != "SENT" {
		t.Fatalf("expected SENT log, got %s", repo.emailLogs[0].Status)
	}
}
