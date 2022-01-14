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
	baseURL    = "https://pro-api.coinmarketcap.com/v1/"
	listingURL = "cryptocurrency/listings/latest"
)

type Crypto struct {
	ID        int     `bson:"_id" json:"_id"`
	Name      string  `bson:"name" json:"name"`
	Symbol    string  `bson:"symbol" json:"symbol"`
	Price     float64 `bson:"price" json:"price"`
	CreatedAt string  `bson:"created_at" json:"created_at"`
}

func convertCryptoToInvesting(data []responses.CryptoPrice, investingList []interface{}) {
	for i, crypto := range data {
		investingList[i] = createInvestingObject(
			createInvestingIDObject(crypto.Symbol, "crypto"),
			crypto.Name,
			crypto.Price.USD.Price,
		)
	}
}

func GetCryptocurrencyPrices() []responses.CryptoPrice {
	url := baseURL + listingURL + "?CMC_PRO_API_KEY=" + os.Getenv("CMC_KEY") + "&limit=300"

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("error: %w", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}

	var data responses.CryptoData
	json.Unmarshal(body, &data)

	return data.Data
}
