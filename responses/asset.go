package responses

type Asset struct {
	ToAsset         string  `bson:"to_asset" json:"to_asset"`
	FromAsset       string  `bson:"from_asset" json:"from_asset"`
	Name            string  `bson:"name" json:"name"`
	RemainingAmount float64 `bson:"remaining_amount" json:"remaining_amount"`
	AssetType       string  `bson:"asset_type" json:"asset_type"`
	TotalValue      float64 `bson:"total_value" json:"total_value"`
	SoldValue       float64 `bson:"sold_value" json:"sold_value"`
	PL              float64 `bson:"p/l" json:"p/l"`
	CurrentPrice    float64 `bson:"current_price" json:"current_price"`
}

type AssetStats struct {
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
