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
	MethodID string `json:"method_id" form:"method_id" binding:"required"`
	Type     *int64 `json:"type" form:"type" binding:"required"`
}

type TransactionUpdate struct {
	ID                 string             `json:"id" binding:"required"`
	Title              *string            `json:"title"`
	Description        *string            `json:"description"`
	Category           *int64             `json:"category"`
	Price              *float64           `json:"price"`
	Currency           *string            `json:"currency"`
	TransactionMethod  *TransactionMethod `json:"method"`
	ShouldDeleteMethod *bool              `json:"delete_method"`
	TransactionDate    *time.Time         `json:"transaction_date"`
}

type TransactionSortFilter struct {
	Category  *int       `form:"category"`
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"end_date" time_format:"2006-01-02"`
	BankAccID *string    `form:"bank_id"`
	CardID    *string    `form:"card_id"`
	Page      int64      `form:"page" binding:"required,number,min=1"`
	Sort      string     `form:"sort" binding:"required,oneof=price date"`
	SortType  int        `form:"type" json:"type" binding:"required,oneof=1 -1"`
}

type TransactionTotalInterval struct {
	Interval        string    `form:"interval" binding:"required,oneof=day month"`
	TransactionDate time.Time `form:"transaction_date" binding:"required" time_format:"2006-01-02"`
}

type TransactionStatsInterval struct {
	Interval string `form:"interval" binding:"required,oneof=weekly monthly yearly"`
}
