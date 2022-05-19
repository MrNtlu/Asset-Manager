package requests

import "time"

type TransactionCreate struct {
	Title             string             `json:"title" binding:"required"`
	Description       *string            `json:"description"`
	Category          *int64             `json:"category" binding:"required"`
	Price             float64            `json:"price" binding:"required"`
	Currency          string             `json:"currency" binding:"required"`
	TransactionMethod *TransactionMethod `json:"method"`
	TransactionDate   time.Time          `json:"transaction_date" binding:"required" time_format:"2006-01-02"`
}

type TransactionMethod struct {
	MethodID string `json:"method_id" binding:"required"`
	Type     *int64 `json:"type" binding:"required"`
}

type TransactionUpdate struct {
	ID                string             `json:"id" binding:"required"`
	Title             *string            `json:"title"`
	Description       *string            `json:"description"`
	Category          *int64             `json:"category"`
	Price             *float64           `json:"price"`
	Currency          *string            `json:"currency"`
	TransactionMethod *TransactionMethod `json:"method"`
	TransactionDate   *time.Time         `json:"transaction_date"`
}
