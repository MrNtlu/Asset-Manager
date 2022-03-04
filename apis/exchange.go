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

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type Exchange struct {
	FromCurrency string  `json:"from_exchange" bson:"from_exchange"`
	ToCurrency   string  `json:"to_exchange" bson:"to_exchange"`
	ExchangeRate float64 `json:"exchange_rate" bson:"exchange_rate"`
}

var baseAlphaURL = "https://www.alphavantage.co/"

func GetExchangeRates() {
	var exchangeDataList []interface{}

	for i := 0; i < len(exchangeList); i++ {
		for j := 0; j < len(exchangeList); j++ {
			if i != j {
				url := baseAlphaURL +
					fmt.Sprintf(
						"query?function=CURRENCY_EXCHANGE_RATE&from_currency=%v&to_currency=%s",
						exchangeList[i],
						exchangeList[j],
					) +
					"&apikey=" + os.Getenv("ALPHAAVANTAGE_KEY")

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

				exchangeRate, _ := strconv.ParseFloat(data.Data.ExchangeRate, 64)
				exchangeDataList = append(exchangeDataList, Exchange{
					FromCurrency: data.Data.FromCurrency,
					ToCurrency:   data.Data.ToCurrency,
					ExchangeRate: exchangeRate,
				})
				time.Sleep(13 * time.Second)
			}
		}
	}
	deleteExchanges()

	if _, err := db.ExchangeCollection.InsertMany(context.TODO(), exchangeDataList); err != nil {
		logrus.Error("failed to create exchange list: ", err)
	}
}

func deleteExchanges() {
	if _, err := db.ExchangeCollection.DeleteMany(context.TODO(), bson.M{}); err != nil {
		logrus.Error("failed to delete exchanges: ", err)
	}
}
