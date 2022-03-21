package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DailyAssetStatsCalculation struct {
	Currency    string             `bson:"currency" json:"currency"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	TotalAssets float64            `bson:"total_assets" json:"total_assets"`
	StockPL     float64            `bson:"stock_p/l" json:"stock_p/l"`
	CryptoPL    float64            `bson:"crypto_p/l" json:"crypto_p/l"`
	ExchangePL  float64            `bson:"exchange_p/l" json:"exchange_p/l"`
	CommodityPL float64            `bson:"commodity_p/l" json:"commodity_p/l"`
	TotalPL     float64            `bson:"total_p/l" json:"total_p/l"`
}

type DailyAssetStats struct {
	Currency    string    `bson:"currency" json:"currency"`
	TotalAssets []float64 `bson:"total_assets" json:"total_assets"`
	TotalPL     []float64 `bson:"total_p/l" json:"total_p/l"`
	StockPL     []float64 `bson:"stock_p/l" json:"stock_p/l"`
	CryptoPL    []float64 `bson:"crypto_p/l" json:"crypto_p/l"`
	ExchangePL  []float64 `bson:"exchange_p/l" json:"exchange_p/l"`
	CommodityPL []float64 `bson:"commodity_p/l" json:"commodity_p/l"`
}
