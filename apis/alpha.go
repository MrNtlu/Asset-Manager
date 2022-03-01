package apis

var exchangeList = []string{"EUR", "JPY", "KRW", "GBP", "USD"}

func convertExchangeToInvesting(investingList []interface{}, size int) {
	for i, exchange := range exchangeList {
		investingList[size+i] = createInvestingObject(
			createInvestingIDObject(exchange, "exchange"),
			exchange,
			1,
		)
	}
}
