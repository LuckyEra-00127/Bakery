package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	CreateProduct(ctx context.Context, product domain.Product) error
	ListProducts(ctx context.Context) ([]domain.Product, error)
	CreateBakePlan(ctx context.Context, plan domain.BakePlan) error
	ListBakePlansByDate(ctx context.Context, date string) ([]domain.BakePlan, error)
	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	ListClientOrders(ctx context.Context, clientID string) ([]domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id, status string) (domain.Order, error)
}

type Publisher interface {
	PublishJSON(subject string, value any)
}

type SalesUseCase struct {
	repo Repository
	pub  Publisher
}

func NewSalesUseCase(repo Repository, pub Publisher) *SalesUseCase {
	return &SalesUseCase{repo: repo, pub: pub}
}

func (uc *SalesUseCase) CreateProduct(ctx context.Context, name string, price float64) (domain.Product, error) {
	if name == "" || price < 0 {
		return domain.Product{}, errors.New("name is required and price must be non-negative")
	}
	product := domain.Product{ID: uuid.NewString(), Name: name, Price: price, CreatedAt: time.Now().UTC()}
	return product, uc.repo.CreateProduct(ctx, product)
}

func (uc *SalesUseCase) ListProducts(ctx context.Context) ([]domain.Product, error) {
	return uc.repo.ListProducts(ctx)
}

func (uc *SalesUseCase) CreateBakePlan(ctx context.Context, productID, date string, planned int) (domain.BakePlan, error) {
	if productID == "" || date == "" || planned < 0 {
		return domain.BakePlan{}, errors.New("product_id, plan_date and planned_quantity are required")
	}
	plan := domain.BakePlan{ID: uuid.NewString(), ProductID: productID, PlanDate: date, PlannedQuantity: planned, AvailableQuantity: planned}
	return plan, uc.repo.CreateBakePlan(ctx, plan)
}

func (uc *SalesUseCase) ListBakePlansByDate(ctx context.Context, date string) ([]domain.BakePlan, error) {
	return uc.repo.ListBakePlansByDate(ctx, date)
}

func (uc *SalesUseCase) CreateOrder(ctx context.Context, clientID, storeName, bakePlanID string, quantity int) (domain.Order, error) {
	if clientID == "" || storeName == "" || bakePlanID == "" || quantity <= 0 {
		return domain.Order{}, errors.New("client_id, store_name, bake_plan_id and positive quantity are required")
	}
	order := domain.Order{
		ClientID:  clientID,
		StoreName: storeName,
		Items: []domain.OrderItem{{
			BakePlanID: bakePlanID,
			Quantity:   quantity,
		}},
	}
	created, err := uc.repo.CreateOrder(ctx, order)
	if err != nil {
		return domain.Order{}, err
	}
	item := created.Items[0]
	uc.pub.PublishJSON("order.created", map[string]any{
		"order_id":     created.ID,
		"client_id":    created.ClientID,
		"store_name":   created.StoreName,
		"product_id":   item.ProductID,
		"bake_plan_id": item.BakePlanID,
		"quantity":     quantity,
		"unit_price":   item.UnitPrice,
		"created_at":   created.CreatedAt,
	})
	return created, nil
}

func (uc *SalesUseCase) ListClientOrders(ctx context.Context, clientID string) ([]domain.Order, error) {
	return uc.repo.ListClientOrders(ctx, clientID)
}

func (uc *SalesUseCase) UpdateOrderStatus(ctx context.Context, id, nextStatus string) (domain.Order, error) {
	id = strings.TrimSpace(id)
	nextStatus = strings.ToUpper(strings.TrimSpace(nextStatus))
	if id == "" {
		return domain.Order{}, errors.New("order id is required")
	}
	switch nextStatus {
	case "PENDING", "CONFIRMED", "DELIVERED", "CANCELLED":
	default:
		return domain.Order{}, errors.New("status must be PENDING, CONFIRMED, DELIVERED or CANCELLED")
	}
	updated, err := uc.repo.UpdateOrderStatus(ctx, id, nextStatus)
	if err != nil {
		return domain.Order{}, err
	}
	uc.pub.PublishJSON("order.status.updated", map[string]any{
		"order_id": updated.ID,
		"status":   updated.Status,
	})
	return updated, nil
}

func (uc *SalesUseCase) CalculateRevenue(soldQuantity int, price float64) float64 {
	return float64(soldQuantity) * price
}
