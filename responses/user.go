package responses

type IsUserPremium struct {
	IsPremium bool `bson:"is_premium" json:"is_premium"`
}
