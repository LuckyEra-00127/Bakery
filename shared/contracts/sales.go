package contracts

type Product struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at,omitempty"`
}

type BakePlan struct {
	ID                string `json:"id"`
	ProductID         string `json:"product_id"`
	ProductName       string `json:"product_name,omitempty"`
	PlanDate          string `json:"plan_date"`
	PlannedQuantity   int    `json:"planned_quantity"`
	AvailableQuantity int    `json:"available_quantity"`
}

type Order struct {
	ID        string      `json:"id"`
	ClientID  string      `json:"client_id"`
	StoreName string      `json:"store_name"`
	Status    string      `json:"status"`
	Items     []OrderItem `json:"items"`
	CreatedAt string      `json:"created_at,omitempty"`
}

type OrderItem struct {
	ID         string  `json:"id"`
	OrderID    string  `json:"order_id"`
	BakePlanID string  `json:"bake_plan_id"`
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
}

type CreateProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type UpdateProductRequest struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type DeleteProductRequest struct {
	ID string `json:"id"`
}
type GetProductByIDRequest struct {
	ID string `json:"id"`
}
type ListProductsRequest struct{}

type ProductResponse struct {
	Product Product `json:"product"`
}

type ListProductsResponse struct {
	Products []Product `json:"products"`
}

type CreateBakePlanRequest struct {
	ProductID       string `json:"product_id"`
	PlanDate        string `json:"plan_date"`
	PlannedQuantity int    `json:"planned_quantity"`
}

type UpdateBakePlanRequest struct {
	ID              string `json:"id"`
	ProductID       string `json:"product_id"`
	PlanDate        string `json:"plan_date"`
	PlannedQuantity int    `json:"planned_quantity"`
}

type DeleteBakePlanRequest struct {
	ID string `json:"id"`
}

type ListBakePlansByDateRequest struct {
	Date string `json:"date"`
}

type GetAvailableProductsRequest struct {
	Date string `json:"date"`
}

type BakePlanResponse struct {
	BakePlan BakePlan `json:"bake_plan"`
}

type ListBakePlansResponse struct {
	BakePlans []BakePlan `json:"bake_plans"`
}

type CreateOrderRequest struct {
	ClientID   string `json:"client_id"`
	StoreName  string `json:"store_name"`
	BakePlanID string `json:"bake_plan_id"`
	Quantity   int    `json:"quantity"`
}

type CancelOrderRequest struct {
	ID string `json:"id"`
}
type GetOrderByIDRequest struct {
	ID string `json:"id"`
}

// UpdateOrderStatusRequest lets an admin move an order through PENDING, CONFIRMED, DELIVERED, or CANCELLED.
type UpdateOrderStatusRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ListClientOrdersRequest struct {
	ClientID string `json:"client_id"`
}

type ListAdminOrdersRequest struct{}

type OrderResponse struct {
	Order Order `json:"order"`
}

type ListOrdersResponse struct {
	Orders []Order `json:"orders"`
}

type CreateDeliveryRequest struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type GetDeliveryByIDRequest struct {
	ID string `json:"id"`
}
type ListDeliveriesRequest struct{}

type CreateReturnReportRequest struct {
	DeliveryID string `json:"delivery_id"`
	Sold       int    `json:"sold"`
	Returned   int    `json:"returned"`
	Replaced   int    `json:"replaced"`
}

type UpdateReturnReportRequest struct {
	ID       string `json:"id"`
	Sold     int    `json:"sold"`
	Returned int    `json:"returned"`
	Replaced int    `json:"replaced"`
}

type ListReturnReportsRequest struct{}

type CalculateRevenueRequest struct {
	SoldQuantity int     `json:"sold_quantity"`
	ProductPrice float64 `json:"product_price"`
}

type CalculateRevenueResponse struct {
	Revenue float64 `json:"revenue"`
}
