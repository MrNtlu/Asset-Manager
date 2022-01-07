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
)

var (
	baseURL    = "https://pro-api.coinmarketcap.com/v1/"
	cryptoURL  = "cryptocurrency/map"
	listingURL = "cryptocurrency/listings/latest"
)

type Crypto struct {
	ID     int     `bson:"id" json:"id"`
	Name   string  `bson:"name" json:"name"`
	Symbol string  `bson:"symbol" json:"symbol"`
	Price  float64 `bson:"price" json:"price"`
}

//TODO: Find a way to convert Crypto prices from USD to whatever currency.
// Decide fixed Currencies like USD GBP EUR TRY AUD KRW JPY etc.
// https://coinmarketcap.com/api/documentation/v1/#tag/cryptocurrency

//TODO: READ https://medium.com/trendyol-tech/concurrency-and-channels-in-go-bbc4dea75286

func createCryptoObject(id int, name, symbol string, price float64) *Crypto {
	return &Crypto{
		ID:     id,
		Name:   name,
		Symbol: symbol,
		Price:  price,
	}
}

func createCrypto(data []responses.CryptoPrice) error {
	cryptoList := make([]interface{}, len(data))
	for i, crypto := range data {
		cryptoList[i] = createCryptoObject(crypto.ID, crypto.Name, crypto.Symbol, crypto.Price.USD.Price)
	}

	if _, err := db.CryptoCollection.InsertMany(context.TODO(), cryptoList); err != nil {
		return fmt.Errorf("failed to create crypto list: %w", err)
	}

	return nil
}

func GetCryptocurrencyPrices() {
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

	createCrypto(data.Data)
}
