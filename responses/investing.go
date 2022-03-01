package responses

type InvestingResponse struct {
	Name   string `bson:"name" json:"name"`
	Symbol string `bson:"symbol" json:"symbol"`
}
