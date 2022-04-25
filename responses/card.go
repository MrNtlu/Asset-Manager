package responses

type CardStatistics struct {
	Currency            string  `bson:"currency" json:"currency"`
	TotalMonthlyPayment float64 `bson:"total_monthly_payment" json:"total_monthly_payment"`
	TotalPayment        float64 `bson:"total_payment" json:"total_payment"`
}
