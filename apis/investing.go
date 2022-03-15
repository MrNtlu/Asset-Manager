package apis

import (
	"asset_backend/db"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	arraySize := len(cryptoList) + len(exchangeList)
	investingList := make([]interface{}, arraySize)

	convertCryptoToInvesting(cryptoList, investingList)
	convertExchangeToInvesting(investingList, len(cryptoList))

	deleteCryptoAndExchange()

	if _, err := db.InvestingCollection.InsertMany(context.TODO(), investingList, options.InsertMany().SetOrdered(false)); err != nil {
		logrus.Error("failed to create investing list: ", err)
	}
}

func deleteCryptoAndExchange() {
	if _, err := db.InvestingCollection.DeleteMany(context.TODO(), bson.M{
		"_id.type": bson.M{
			"$in": bson.A{"crypto", "exchange"},
		},
	}); err != nil {
		logrus.Error("failed to delete exchange and crypto: ", err)
	}
}
