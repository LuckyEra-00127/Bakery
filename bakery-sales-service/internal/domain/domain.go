package domain

import "time"

type Product struct {
	ID        string
	Name      string
	Price     float64
	CreatedAt time.Time
}

type BakePlan struct {
	ID                string
	ProductID         string
	ProductName       string
	PlanDate          string
	PlannedQuantity   int
	AvailableQuantity int
}

type Order struct {
	ID        string
	ClientID  string
	StoreName string
	Status    string
	Items     []OrderItem
	CreatedAt time.Time
}

type OrderItem struct {
	ID         string
	OrderID    string
	BakePlanID string
	ProductID  string
	Quantity   int
	UnitPrice  float64
}
