package responses

type Asset struct {
	ToAsset         string  `bson:"to_asset" json:"to_asset"`
	FromAsset       string  `bson:"from_asset" json:"from_asset"`
	RemainingAmount float64 `bson:"remaining_amount" json:"remaining_amount"`
	AssetType       string  `bson:"asset_type" json:"asset_type"`
	TotalValue      float64 `bson:"total_value" json:"total_value"`
	SoldValue       float64 `bson:"sold_value" json:"sold_value"`
	PL              float64 `bson:"p/l" json:"p/l"`
	CurrentPrice    float64 `bson:"current_price" json:"current_price"`
}
