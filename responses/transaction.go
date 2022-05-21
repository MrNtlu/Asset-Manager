package responses

import "time"

type TransactionCalendarCount struct {
	ID    time.Time `bson:"_id" json:"_id"`
	Count int       `bson:"count" json:"count"`
}

type TransactionTotal struct {
	Currency         string  `bson:"currency" json:"currency"`
	TotalTransaction float64 `bson:"total_transaction" json:"total_transaction"`
}

type TransactionStats struct {
	Currency         string    `bson:"currency" json:"currency"`
	TotalTransaction float64   `bson:"total_transaction" json:"total_transaction"`
	Date             time.Time `bson:"date" json:"date"`
}
