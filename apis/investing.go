package apis

import (
	"asset_backend/db"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Investing struct {
	ID        InvestingID `bson:"_id" json:"_id"`
	Name      string      `bson:"name" json:"name"`
	Price     float64     `bson:"price" json:"price"`
	CreatedAt string      `bson:"created_at" json:"created_at"`
}

type InvestingID struct {
	Symbol string `bson:"symbol" json:"symbol"`
	Type   string `bson:"type" json:"type"`
}

func createInvestingObject(id InvestingID, name string, price float64) *Investing {
	return &Investing{
		ID:        id,
		Name:      name,
		Price:     price,
		CreatedAt: time.Now().Format(time.RFC822),
	}
}

func createInvestingIDObject(symbol, tType string) InvestingID {
	return InvestingID{
		Symbol: symbol,
		Type:   tType,
	}
}

func GetAndCreateInvesting() {
	cryptoList := GetCryptocurrencyPrices()
	exchangeList := GetExchangeRates()
	stockList := GetStockPrices()

	arraySize := len(cryptoList) + len(exchangeList) + len(stockList)
	investingList := make([]interface{}, arraySize)

	convertCryptoToInvesting(cryptoList, investingList)
	convertExchangeToInvesting(exchangeList, investingList, len(cryptoList))
	convertStockToInvesting(stockList, investingList, (len(cryptoList) + len(exchangeList)))

	deleteInvestings()

	if _, err := db.InvestingCollections.InsertMany(context.TODO(), investingList); err != nil {
		fmt.Println("failed to create investing list: %w", err)
	}
}

func deleteInvestings() {
	_, err := db.InvestingCollections.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println("error: %w", err)
	}
}
