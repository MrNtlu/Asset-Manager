package responses

type InvestingResponse struct {
	Name   string `bson:"name" json:"name"`
	Symbol string `bson:"symbol" json:"symbol"`
}

type InvestingTableResponse struct {
	Name     string  `bson:"name" json:"name"`
	Symbol   string  `bson:"symbol" json:"symbol"`
	Price    float64 `bson:"price" json:"price"`
	Market   string  `bson:"market" json:"market"`
	Currency string  `bson:"currency" json:"currency"`
}
