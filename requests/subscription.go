package requests

import "time"

type Subscription struct {
	CardID      *string    `json:"card_id"`
	Name        string     `json:"name" binding:"required"`
	Description *string    `json:"description"`
	BillDate    time.Time  `json:"bill_date" binding:"required" time_format:"2006-01-02"`
	BillCycle   *BillCycle `json:"bill_cycle"`
	Price       float32    `json:"price" binding:"required"`
	Currency    string     `json:"currency" binding:"required"`
}

type BillCycle struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

type Card struct {
	Name       string `json:"name" binding:"required"`
	Last4Digit string `json:"last_digit" binding:"required"`
}

type SubscriptionUpdate struct {
	SubscriptionID string     `json:"subscription_id" binding:"required"`
	Name           *string    `json:"name"`
	Description    *string    `json:"description"`
	BillDate       *time.Time `json:"bill_date"`
	BillCycle      *BillCycle `json:"bill_cycle"`
	Price          *float32   `json:"price"`
	Currency       *string    `json:"currency"`
}
