package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/domain"
	"github.com/google/uuid"
)

type SalesRepository struct {
	db *sql.DB
}

func NewSalesRepository(db *sql.DB) *SalesRepository {
	return &SalesRepository{db: db}
}

func (r *SalesRepository) CreateProduct(ctx context.Context, product domain.Product) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO products (id, name, price, created_at) VALUES ($1, $2, $3, $4)`, product.ID, product.Name, product.Price, product.CreatedAt)
	return err
}

func (r *SalesRepository) ListProducts(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, price, created_at FROM products ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *SalesRepository) CreateBakePlan(ctx context.Context, plan domain.BakePlan) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO bake_plans (id, product_id, plan_date, planned_quantity, available_quantity)
        VALUES ($1, $2, $3, $4, $5)
    `, plan.ID, plan.ProductID, plan.PlanDate, plan.PlannedQuantity, plan.AvailableQuantity)
	return err
}

func (r *SalesRepository) ListBakePlansByDate(ctx context.Context, date string) ([]domain.BakePlan, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT bp.id, bp.product_id, p.name, bp.plan_date::text, bp.planned_quantity, bp.available_quantity
        FROM bake_plans bp
        JOIN products p ON p.id = bp.product_id
        WHERE ($1 = '' OR bp.plan_date = $1::date)
        ORDER BY bp.plan_date ASC, p.name ASC
    `, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []domain.BakePlan
	for rows.Next() {
		var p domain.BakePlan
		if err := rows.Scan(&p.ID, &p.ProductID, &p.ProductName, &p.PlanDate, &p.PlannedQuantity, &p.AvailableQuantity); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *SalesRepository) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Order{}, err
	}
	defer tx.Rollback()

	if len(order.Items) != 1 {
		return domain.Order{}, errors.New("this endpoint currently supports one order item per order")
	}
	item := order.Items[0]

	var productID string
	var unitPrice float64
	var available int
	err = tx.QueryRowContext(ctx, `
        SELECT bp.product_id, p.price, bp.available_quantity
        FROM bake_plans bp
        JOIN products p ON p.id = bp.product_id
        WHERE bp.id = $1
        FOR UPDATE
    `, item.BakePlanID).Scan(&productID, &unitPrice, &available)
	if err != nil {
		return domain.Order{}, err
	}
	if available < item.Quantity {
		return domain.Order{}, errors.New("not enough available quantity")
	}

	order.ID = uuid.NewString()
	order.Status = "PENDING"
	order.CreatedAt = time.Now().UTC()
	item.ID = uuid.NewString()
	item.OrderID = order.ID
	item.ProductID = productID
	item.UnitPrice = unitPrice

	_, err = tx.ExecContext(ctx, `INSERT INTO orders (id, client_id, store_name, status, created_at) VALUES ($1, $2, $3, $4, $5)`, order.ID, order.ClientID, order.StoreName, order.Status, order.CreatedAt)
	if err != nil {
		return domain.Order{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO order_items (id, order_id, bake_plan_id, product_id, quantity, unit_price) VALUES ($1, $2, $3, $4, $5, $6)`, item.ID, item.OrderID, item.BakePlanID, item.ProductID, item.Quantity, item.UnitPrice)
	if err != nil {
		return domain.Order{}, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE bake_plans SET available_quantity = available_quantity - $1 WHERE id = $2`, item.Quantity, item.BakePlanID)
	if err != nil {
		return domain.Order{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Order{}, err
	}

	order.Items = []domain.OrderItem{item}
	return order, nil
}

func (r *SalesRepository) ListClientOrders(ctx context.Context, clientID string) ([]domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, client_id, store_name, status, created_at
        FROM orders
        WHERE ($1 = '' OR client_id = $1)
        ORDER BY created_at DESC
    `, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.ClientID, &o.StoreName, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Items, err = r.listOrderItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *SalesRepository) UpdateOrderStatus(ctx context.Context, id, nextStatus string) (domain.Order, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE orders
		SET status = $2
		WHERE id = $1
		RETURNING id, client_id, store_name, status, created_at
	`, id, nextStatus)

	var order domain.Order
	if err := row.Scan(&order.ID, &order.ClientID, &order.StoreName, &order.Status, &order.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Order{}, errors.New("order not found")
		}
		return domain.Order{}, err
	}

	items, err := r.listOrderItems(ctx, order.ID)
	if err != nil {
		return domain.Order{}, err
	}
	order.Items = items
	return order, nil
}

func (r *SalesRepository) listOrderItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, order_id, bake_plan_id, product_id, quantity, unit_price FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.BakePlanID, &item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
