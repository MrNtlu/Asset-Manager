package apis

import (
	"asset_backend/apis/responses"
	"asset_backend/db"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	baseAlphaURL = "https://www.alphavantage.co/"
	currencyURL  = "query?function=CURRENCY_EXCHANGE_RATE&from_currency=USD&to_currency="
	stockURL     = ""

	exchangeList = []string{"EUR", "JPY", "KRW", "GBP"}
)

type Exchange struct {
	ID           string    `bson:"_id" json:"_id"`
	ExchangeRate float64   `bson:"exchange_rate" json:"exchange_rate"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}

func createExchangeObject(id string, rate float64) *Exchange {
	return &Exchange{
		ID:           id,
		ExchangeRate: rate,
		CreatedAt:    time.Now(),
	}
}

func createExchange(data []responses.ExchangeData) error {
	exchangeList := make([]interface{}, len(data))
	for i, exchange := range data {
		exchangeRate, _ := strconv.ParseFloat(exchange.Data.ExchangeRate, 64)
		exchangeList[i] = createExchangeObject(exchange.Data.ToCurrency, exchangeRate)
	}

	if _, err := db.ExchangeCollection.InsertMany(context.TODO(), exchangeList); err != nil {
		return fmt.Errorf("failed to create exchange list: %w", err)
	}

	return nil
}

func GetExchangeRates() {
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

	createExchange(exchangeDataList)
}
