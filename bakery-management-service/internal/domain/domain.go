package domain

import "time"

type IngredientSection struct {
	Name   string
	Amount string
	Unit   string
}

type Ingredient struct {
	ID          string
	Name        string
	Unit        string
	CostPerUnit float64
	Sections    []IngredientSection
}

type Task struct {
	ID        string
	Title     string
	Status    string
	DueDate   string
	CreatedAt time.Time
}

type DailyStatistic struct {
	Date      string
	ProductID string
	Product   string
	Baked     int
	Delivered int
	Sold      int
	Returned  int
	Left      int
	Revenue   float64
}

type EmailLog struct {
	ID        string
	Recipient string
	Subject   string
	Body      string
	Status    string
	CreatedAt time.Time
}
