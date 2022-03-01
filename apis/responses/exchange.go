package responses

type ExchangeData struct {
	Data ExchangeRate `json:"Realtime Currency Exchange Rate"`
}

type ExchangeRate struct {
	FromCurrency string `json:"1. From_Currency Code"`
	ToCurrency   string `json:"3. To_Currency Code"`
	ExchangeRate string `json:"5. Exchange Rate"`
}
