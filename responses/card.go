package responses

type CardStatistics struct {
	Currency            string  `bson:"currency" json:"currency"`
	CardName            string  `bson:"card_name" json:"card_name"`
	CardLastDigit       string  `bson:"card_last_digit" json:"card_last_digit"`
	MEName              string  `bson:"most_expensive_name" json:"most_expensive_name"`
	ME                  float64 `bson:"most_expensive" json:"most_expensive"`
	TotalMonthlyPayment float64 `bson:"total_monthly_payment" json:"total_monthly_payment"`
	TotalPayment        float64 `bson:"total_payment" json:"total_payment"`
}
