package requests

type AssetCreate struct {
	ToAsset     string   `json:"to_asset" binding:"required"`
	FromAsset   string   `json:"from_asset" binding:"required"`
	BoughtPrice *float64 `json:"bought_price"`
	SoldPrice   *float64 `json:"sold_price"`
	Amount      float64  `json:"amount" binding:"required"`
	AssetType   string   `json:"asset_type" binding:"required,oneof=crypto currency exchange"`
	Type        string   `json:"type" binding:"required,oneof=sell buy"`
}

type AssetSort struct {
	Sort string `json:"sort" binding:"required,oneof=name worth amount worthless"`
}

type AssetLog struct {
	ToAsset   string  `json:"to_asset" binding:"required"`
	FromAsset string  `json:"from_asset" binding:"required"`
	Type      *string `json:"type" binding:"oneof=sell buy"`
	Sort      string  `json:"sort" binding:"required,oneof=newest oldest amount"`
	Page      int64   `json:"page" binding:"required,number,min=1"`
}

type AssetUpdate struct {
	AssetID     string   `json:"asset_id" binding:"required"`
	BoughtPrice *float64 `json:"bought_price"`
	SoldPrice   *float64 `json:"sold_price"`
	Amount      float64  `json:"amount"`
}

type AssetLogsDelete struct {
	ToAsset   string `json:"to_asset" binding:"required"`
	FromAsset string `json:"from_asset" binding:"required"`
}
