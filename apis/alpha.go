package apis

import (
	"asset_backend/apis/responses"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

var (
	baseAlphaURL = "https://www.alphavantage.co/"
	currencyURL  = "query?function=CURRENCY_EXCHANGE_RATE&from_currency=USD&to_currency="

	exchangeList = []string{"EUR", "JPY", "KRW", "GBP", "USD"}
)

type Exchange struct {
	ID           string  `bson:"_id" json:"_id"`
	ExchangeRate float64 `bson:"exchange_rate" json:"exchange_rate"`
	CreatedAt    string  `bson:"created_at" json:"created_at"`
}

func convertExchangeToInvesting(data []responses.ExchangeData, investingList []interface{}, size int) {
	for i, exchange := range data {
		exchangeRate, _ := strconv.ParseFloat(exchange.Data.ExchangeRate, 64)

		investingList[size+i] = createInvestingObject(
			createInvestingIDObject(exchange.Data.ToCurrency, "exchange"),
			exchange.Data.ToCurrency,
			exchangeRate,
		)
	}
}

func GetExchangeRates() []responses.ExchangeData {
	var exchangeDataList []responses.ExchangeData
	for _, exchange := range exchangeList {
		url := baseAlphaURL + currencyURL + exchange + "&apikey=" + os.Getenv("ALPHAAVANTAGE_KEY")

		response, err := http.Get(url)
		if err != nil {
			fmt.Println("error: %w", err)
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err.Error())
		}

		var data responses.ExchangeData
		json.Unmarshal(body, &data)

		exchangeDataList = append(exchangeDataList, data)
	}

	return exchangeDataList
}
