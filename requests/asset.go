package requests

type Asset struct {
	ToAsset     string   `json:"to_asset" binding:"required"`
	FromAsset   string   `json:"from_asset" binding:"required"`
	BoughtPrice float64  `json:"bought_price" binding:"required"`
	SoldPrice   *float64 `json:"sold_price"`
	Amount      float64  `json:"amount" binding:"required"`
	AssetType   string   `json:"asset_type" binding:"required,oneof=crypto currency exchange"`
	Type        string   `json:"type" binding:"required,oneof=sell buy"`
}
