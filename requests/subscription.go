package requests

import "time"

type Subscription struct {
	CardID      *string   `json:"card_id"`
	Name        string    `json:"name" binding:"required"`
	Description *string   `json:"description"`
	BillDate    time.Time `json:"bill_date" binding:"required" time_format:"2006-01-02"`
	BillCycle   *int      `json:"bill_cycle"`
	Price       float64   `json:"price" binding:"required"`
	Currency    string    `json:"currency" binding:"required"`
	Color       string    `json:"color" binding:"required"`
	Image       *string   `json:"image"`
}

type SubscriptionUpdate struct {
	ID          string     `json:"id" binding:"required"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	BillDate    *time.Time `json:"bill_date"`
	BillCycle   *int       `json:"bill_cycle"`
	Price       *float64   `json:"price"`
	Currency    *string    `json:"currency"`
	CardID      *string    `json:"card_id"`
	Color       *string    `json:"color"`
	Image       *string    `json:"image"`
}

type SubscriptionSort struct {
	Sort     string `form:"sort" binding:"required,oneof=name currency price"`
	SortType int    `form:"type" binding:"required,oneof=1 -1"`
}
