package apis

import "time"

var (
	baseMarketStackURL = "http://api.marketstack.com/v1/"
	tickersURL         = "tickers"
)

type Stock struct {
	ID        int       `bson:"_id" json:"_id"`
	Name      string    `bson:"name" json:"name"`
	Symbol    string    `bson:"symbol" json:"symbol"`
	Price     float64   `bson:"price" json:"price"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
