package responses

type IsUserPremium struct {
	IsPremium bool `bson:"is_premium" json:"is_premium"`
}

type UserInfo struct {
	IsPremium         bool   `bson:"is_premium" json:"is_premium"`
	IsLifetimePremium bool   `bson:"is_lifetime_premium" json:"is_lifetime_premium"`
	EmailAddress      string `bson:"email_address" json:"email_address"`
	Currency          string `bson:"currency" json:"currency"`
	InvestingLimit    string `bson:"investing_limit" json:"investing_limit"`
	SubscriptionLimit string `bson:"subscription_limit" json:"subscription_limit"`
}
