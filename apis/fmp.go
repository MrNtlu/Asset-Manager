package apis

import (
	"asset_backend/apis/responses"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	baseFMPURL       = "https://financialmodelingprep.com/api/v3/"
	stockURL         = "stock/list"
	availableListURL = "available-traded/list"
)

type Stock struct {
	ID        string  `bson:"_id" json:"_id"`
	Name      string  `bson:"name" json:"name"`
	Price     float64 `bson:"price" json:"price"`
	CreatedAt string  `bson:"created_at" json:"created_at"`
}

func convertStockToInvesting(data []responses.StockData, investingList []interface{}, size int) {
	for i, stock := range data {
		investingList[size+i] = createInvestingObject(
			createInvestingIDObject(stock.Symbol, "stock"),
			stock.Name,
			stock.Price,
		)
	}
}

func GetStockPrices() []responses.StockData {
	url := baseFMPURL + stockURL + "?apikey=" + os.Getenv("FMP_KEY")

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("error: %w", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}

	var data []responses.StockData
	json.Unmarshal(body, &data)

	return data
}

func GetAvailableStockPrices() []responses.StockData {
	url := baseFMPURL + availableListURL + "?apikey=" + os.Getenv("FMP_KEY")

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("error: %w", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}

	var data []responses.StockData
	json.Unmarshal(body, &data)

	return data
}
