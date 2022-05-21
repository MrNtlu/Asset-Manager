package responses

type CardSubscriptionStatistics struct {
	Currency            string  `bson:"currency" json:"currency"`
	TotalMonthlyPayment float64 `bson:"total_monthly_payment" json:"total_monthly_payment"`
	TotalPayment        float64 `bson:"total_payment" json:"total_payment"`
}

type CardStats struct {
	SubscriptionStats CardSubscriptionStatistics `bson:"subscription_stats" json:"subscription_stats"`
	TransactionStats  TransactionTotal           `bson:"transaction_stats" json:"transaction_stats"`
}
