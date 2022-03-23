package responses

type IsUserPremium struct {
	IsPremium bool `bson:"is_premium" json:"is_premium"`
}

type UserInfo struct {
	IsPremium         bool   `bson:"is_premium" json:"is_premium"`
	InvestingLimit    string `bson:"investing_limit" json:"investing_limit"`
	SubscriptionLimit string `bson:"subscription_limit" json:"subscription_limit"`
}
