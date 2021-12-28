package requests

import "time"

type Subscription struct {
	UserID      string     `json:"user_id" binding:"required"`
	CardID      *string    `json:"card_id"`
	Name        string     `json:"name" binding:"required"`
	Description *string    `json:"description"`
	BillDate    time.Time  `json:"bill_date" binding:"required,billDate" time_format:"2006-01-02"`
	BillCycle   *BillCycle `json:"bill_cycle" binding:"required"`
	Price       float32    `json:"price" binding:"required"`
	Currency    string     `json:"currency" binding:"required"`
}

type BillCycle struct {
	Day   int `bson:"day" json:"day"`
	Month int `bson:"month" json:"month"`
	Year  int `bson:"year" json:"year"`
}

type Card struct {
	UserID     string `json:"user_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Last4Digit string `json:"last_digit" binding:"required"`
	CardType   string `json:"type" binding:"required"`
}
