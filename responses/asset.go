package responses

import "time"

type Asset struct {
	ToAsset         string  `bson:"to_asset" json:"to_asset"`
	FromAsset       string  `bson:"from_asset" json:"from_asset"`
	Name            string  `bson:"name" json:"name"`
	RemainingAmount float64 `bson:"remaining_amount" json:"remaining_amount"`
	AssetType       string  `bson:"asset_type" json:"asset_type"`
	TotalBought     float64 `bson:"total_bought" json:"total_bought"`
	TotalSold       float64 `bson:"total_sold" json:"total_sold"`
	PL              float64 `bson:"p/l" json:"p/l"`
	CurrentTotal    float64 `bson:"current_total_value" json:"current_total_value"`
}

type AssetDetails struct {
	ToAsset         string     `bson:"to_asset" json:"to_asset"`
	FromAsset       string     `bson:"from_asset" json:"from_asset"`
	Name            string     `bson:"name" json:"name"`
	RemainingAmount float64    `bson:"remaining_amount" json:"remaining_amount"`
	TotalBought     float64    `bson:"total_bought" json:"total_bought"`
	TotalSold       float64    `bson:"total_sold" json:"total_sold"`
	CurrentTotal    float64    `bson:"current_total_value" json:"current_total_value"`
	PL              float64    `bson:"p/l" json:"p/l"`
	AssetType       string     `bson:"asset_type" json:"asset_type"`
	Assets          []AssetLog `bson:"assets" json:"assets"`
}

type AssetLog struct {
	ToAsset       string    `bson:"to_asset" json:"to_asset"`
	FromAsset     string    `bson:"from_asset" json:"from_asset"`
	Price         float64   `bson:"price" json:"price"`
	Amount        float64   `bson:"amount" json:"amount"`
	Type          string    `bson:"type" json:"type"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	CurrencyValue float64   `bson:"value" json:"value"`
}

type AssetStats struct {
	Currency           string  `bson:"currency" json:"currency"`
	TotalBought        float64 `bson:"total_bought" json:"total_bought"`
	TotalSold          float64 `bson:"total_sold" json:"total_sold"`
	StockAssets        float64 `bson:"stock_assets" json:"stock_assets"`
	CryptoAssets       float64 `bson:"crypto_assets" json:"crypto_assets"`
	ExchangeAssets     float64 `bson:"exchange_assets" json:"exchange_assets"`
	TotalAssets        float64 `bson:"total_assets" json:"total_assets"`
	StockPL            float64 `bson:"stock_p/l" json:"stock_p/l"`
	CryptoPL           float64 `bson:"crypto_p/l" json:"crypto_p/l"`
	ExchangePL         float64 `bson:"exchange_p/l" json:"exchange_p/l"`
	TotalPL            float64 `bson:"total_p/l" json:"total_p/l"`
	StockPercentage    float64 `bson:"stock_percentage" json:"stock_percentage"`
	CryptoPercentage   float64 `bson:"crypto_percentage" json:"crypto_percentage"`
	ExchangePercentage float64 `bson:"exchange_percentage" json:"exchange_percentage"`
}
