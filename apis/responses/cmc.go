package responses

type CryptoData struct {
	Data []CryptoPrice `bson:"data" json:"data"`
}

type CryptoPrice struct {
	ID     int       `bson:"id" json:"id"`
	Name   string    `json:"name"`
	Symbol string    `json:"symbol"`
	Price  CryptoUSD `json:"quote"`
}

type CryptoUSD struct {
	USD CryptoQuote `json:"USD"`
}

type CryptoQuote struct {
	Price float64 `json:"price"`
}
