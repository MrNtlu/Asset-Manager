package responses

type IsUserPremium struct {
	IsPremium         bool `bson:"is_premium" json:"is_premium"`
	IsLifetimePremium bool `bson:"is_lifetime_premium" json:"is_lifetime_premium"`
}

type UserInfo struct {
	IsPremium         bool   `bson:"is_premium" json:"is_premium"`
	IsLifetimePremium bool   `bson:"is_lifetime_premium" json:"is_lifetime_premium"`
	IsOAuth           bool   `bson:"is_oauth" json:"is_oauth"`
	AppNotification   bool   `bson:"app_notification" json:"app_notification"`
	EmailAddress      string `bson:"email_address" json:"email_address"`
	Currency          string `bson:"currency" json:"currency"`
	InvestingLimit    string `bson:"investing_limit" json:"investing_limit"`
	SubscriptionLimit string `bson:"subscription_limit" json:"subscription_limit"`
	WatchlistLimit    string `bson:"watchlist_limit" json:"watchlist_limit"`
	FCMToken          string `bson:"fcm_token" json:"fcm_token"`
}
