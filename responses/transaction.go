package responses

import "time"

type TransactionTotal struct {
	Currency         string  `bson:"currency" json:"currency"`
	TotalTransaction float64 `bson:"total_transaction" json:"total_transaction"`
}

type TransactionStats struct {
	TransactionDailyStats    []TransactionDailyStats  `bson:"daily_stats" json:"daily_stats"`
	TransactionCategoryStats TransactionCategoryStats `bson:"category_stats" json:"category_stats"`
	TotalExpense             float64                  `bson:"total_expense" json:"total_expense"`
	TotalIncome              float64                  `bson:"total_income" json:"total_income"`
}

type TransactionDailyStats struct {
	Currency         string    `bson:"currency" json:"currency"`
	TotalTransaction float64   `bson:"total_transaction" json:"total_transaction"`
	Date             time.Time `bson:"date" json:"date"`
}

type TransactionCategoryStats struct {
	Currency         string                    `bson:"currency" json:"currency"`
	TotalTransaction float64                   `bson:"total_transaction" json:"total_transaction"`
	CategoryList     []TransactionCategoryStat `bson:"category_list" json:"category_list"`
}

type TransactionCategoryStat struct {
	CategoryID               int64   `bson:"_id" json:"_id"`
	TotalCategoryTransaction float64 `bson:"total_transaction" json:"total_transaction"`
}
