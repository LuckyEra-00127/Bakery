package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/domain"
)

type fakeSalesRepo struct {
	products []domain.Product
	orders   map[string]domain.Order
}

func (r *fakeSalesRepo) CreateProduct(_ context.Context, product domain.Product) error {
	r.products = append(r.products, product)
	return nil
}

func (r *fakeSalesRepo) ListProducts(_ context.Context) ([]domain.Product, error) {
	return r.products, nil
}

func (r *fakeSalesRepo) CreateBakePlan(_ context.Context, plan domain.BakePlan) error {
	return nil
}

func (r *fakeSalesRepo) ListBakePlansByDate(_ context.Context, _ string) ([]domain.BakePlan, error) {
	return nil, nil
}

func (r *fakeSalesRepo) CreateOrder(_ context.Context, order domain.Order) (domain.Order, error) {
	order.ID = "order-1"
	order.Status = "PENDING"
	order.CreatedAt = time.Now().UTC()
	order.Items[0].ID = "item-1"
	order.Items[0].ProductID = "product-1"
	order.Items[0].UnitPrice = 200
	if r.orders == nil {
		r.orders = map[string]domain.Order{}
	}
	r.orders[order.ID] = order
	return order, nil
}

func (r *fakeSalesRepo) ListClientOrders(_ context.Context, _ string) ([]domain.Order, error) {
	return nil, nil
}

func (r *fakeSalesRepo) UpdateOrderStatus(_ context.Context, id, status string) (domain.Order, error) {
	if r.orders == nil || r.orders[id].ID == "" {
		return domain.Order{}, errors.New("order not found")
	}
	order := r.orders[id]
	order.Status = status
	r.orders[id] = order
	return order, nil
}

type fakePublisher struct {
	subjects []string
}

func (p *fakePublisher) PublishJSON(subject string, _ any) {
	p.subjects = append(p.subjects, subject)
}

func TestCreateProductValidation(t *testing.T) {
	uc := NewSalesUseCase(&fakeSalesRepo{}, &fakePublisher{})
	if _, err := uc.CreateProduct(context.Background(), "", 100); err == nil {
		t.Fatal("expected validation error for empty product name")
	}
	product, err := uc.CreateProduct(context.Background(), "Bread", 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product.Name != "Bread" || product.Price != 200 {
		t.Fatalf("unexpected product: %+v", product)
	}
}

func TestUpdateOrderStatusValidationAndPublish(t *testing.T) {
	repo := &fakeSalesRepo{orders: map[string]domain.Order{"order-1": {ID: "order-1", Status: "PENDING"}}}
	pub := &fakePublisher{}
	uc := NewSalesUseCase(repo, pub)

	if _, err := uc.UpdateOrderStatus(context.Background(), "order-1", "bad"); err == nil {
		t.Fatal("expected invalid status error")
	}
	updated, err := uc.UpdateOrderStatus(context.Background(), "order-1", "confirmed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != "CONFIRMED" {
		t.Fatalf("expected CONFIRMED, got %s", updated.Status)
	}
	if len(pub.subjects) != 1 || pub.subjects[0] != "order.status.updated" {
		t.Fatalf("expected order.status.updated event, got %+v", pub.subjects)
	}
}
