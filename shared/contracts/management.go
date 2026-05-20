package contracts

type IngredientSection struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

type Ingredient struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Unit        string              `json:"unit"`
	CostPerUnit float64             `json:"cost_per_unit"`
	Sections    []IngredientSection `json:"sections,omitempty"`
}

type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	DueDate   string `json:"due_date,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type DailyStatistic struct {
	Date      string  `json:"date"`
	ProductID string  `json:"product_id"`
	Product   string  `json:"product,omitempty"`
	Baked     int     `json:"baked"`
	Delivered int     `json:"delivered"`
	Sold      int     `json:"sold"`
	Returned  int     `json:"returned"`
	Left      int     `json:"left"`
	Revenue   float64 `json:"revenue"`
}

type CreateIngredientRequest struct {
	Name        string              `json:"name"`
	Unit        string              `json:"unit"`
	CostPerUnit float64             `json:"cost_per_unit"`
	Sections    []IngredientSection `json:"sections,omitempty"`
}

type UpdateIngredientRequest struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Unit        string              `json:"unit"`
	CostPerUnit float64             `json:"cost_per_unit"`
	Sections    []IngredientSection `json:"sections,omitempty"`
}

type DeleteIngredientRequest struct {
	ID string `json:"id"`
}
type ListIngredientsRequest struct{}

type IngredientResponse struct {
	Ingredient Ingredient `json:"ingredient"`
}

type ListIngredientsResponse struct {
	Ingredients []Ingredient `json:"ingredients"`
}

type CreateRecipeRequest struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
}

type UpdateRecipeRequest struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
}

type DeleteRecipeRequest struct {
	ID string `json:"id"`
}
type GetRecipeByProductIDRequest struct {
	ProductID string `json:"product_id"`
}

type AddIngredientToRecipeRequest struct {
	RecipeID     string  `json:"recipe_id"`
	IngredientID string  `json:"ingredient_id"`
	Amount       float64 `json:"amount"`
}

type RemoveIngredientFromRecipeRequest struct {
	RecipeID     string `json:"recipe_id"`
	IngredientID string `json:"ingredient_id"`
}

type CalculateRecipeRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CalculateRecipeCostRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CreateTaskRequest struct {
	Title   string `json:"title"`
	Status  string `json:"status,omitempty"`
	DueDate string `json:"due_date,omitempty"`
}

type UpdateTaskRequest struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	DueDate string `json:"due_date,omitempty"`
}

type DeleteTaskRequest struct {
	ID string `json:"id"`
}
type ListTasksRequest struct{}
type CompleteTaskRequest struct {
	ID string `json:"id"`
}

type TaskResponse struct {
	Task Task `json:"task"`
}

type ListTasksResponse struct {
	Tasks []Task `json:"tasks"`
}

type GetDailyStatisticsRequest struct {
	Date string `json:"date"`
}
type GetWeeklyStatisticsRequest struct {
	Week string `json:"week"`
}
type GetMonthlyStatisticsRequest struct {
	Month string `json:"month"`
}

type DailyStatisticsResponse struct {
	Statistics []DailyStatistic `json:"statistics"`
}

type SendEmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type SendEmailResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type EmailLog struct {
	ID        string `json:"id"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at,omitempty"`
}

type ListEmailLogsRequest struct{}

type ListEmailLogsResponse struct {
	EmailLogs []EmailLog `json:"email_logs"`
}
