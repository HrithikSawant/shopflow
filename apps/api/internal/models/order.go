package models

type Order struct {
	ID          int64
	UserID      string
	Status      string
	TotalAmount float64
}
