package apis

var exchangeList = []string{"EUR", "JPY", "KRW", "GBP", "USD"}
var exchangeNameList = []string{"Euro", "Japanese Yen", "South Korean Won", "British Pound Sterling", "United States Dollar"}

func convertExchangeToInvesting(investingList []interface{}, size int) {
	for i, exchange := range exchangeList {
		investingList[size+i] = createInvestingObject(
			createInvestingIDObject(exchange, "exchange"),
			exchangeNameList[i],
			1,
		)
	}
}
