package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type FavouriteInvesting struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	UserID      string               `bson:"user_id" json:"user_id"`
	InvestingID FavouriteInvestingID `bson:"investing_id" json:"investing_id"`
	Priority    int                  `bson:"priority" json:"priority"`
	Investing   Investing            `bson:"investing" json:"investing"`
}

type Investing struct {
	Price    float64 `bson:"price" json:"price"`
	Currency string  `bson:"currency" json:"currency"`
}

type FavouriteInvestingID struct {
	Symbol string `bson:"symbol" json:"symbol"`
	Type   string `bson:"type" json:"type"`
	Market string `bson:"market" json:"market"`
}
