package responses

type InvestingResponse struct {
	Name   string  `bson:"name" json:"name"`
	Price  float64 `bson:"price" json:"price"`
	Symbol string  `bson:"symbol" json:"symbol"`
}
