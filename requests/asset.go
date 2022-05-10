package requests

type AssetCreate struct {
	ToAsset     string  `json:"to_asset" binding:"required"`
	FromAsset   string  `json:"from_asset" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	AssetType   string  `json:"asset_type" binding:"required,oneof=crypto stock exchange commodity"`
	AssetMarket string  `json:"asset_market"`
	Type        string  `json:"type" binding:"required,oneof=sell buy"`
}

type AssetSort struct {
	Sort     string `form:"sort" binding:"required,oneof=name percentage amount profit"`
	SortType int    `form:"type" json:"type" binding:"required,number,oneof=1 -1"`
}

type AssetDetails struct {
	ToAsset     string `form:"to_asset" json:"to_asset" binding:"required"`
	FromAsset   string `form:"from_asset" json:"from_asset" binding:"required"`
	AssetMarket string `form:"asset_market" json:"asset_market" binding:"required"`
}

type AssetLog struct {
	ToAsset     string `form:"to_asset" json:"to_asset" binding:"required"`
	FromAsset   string `form:"from_asset" json:"from_asset" binding:"required"`
	AssetMarket string `form:"asset_market" json:"asset_market" binding:"required"`
	Sort        string `form:"sort" json:"sort" binding:"required,oneof=newest oldest amount"`
	Page        int64  `form:"page" json:"page" binding:"required,number,min=1"`
}

type AssetUpdate struct {
	ID     string   `json:"id" binding:"required"`
	Type   *string  `json:"type"`
	Price  *float64 `json:"price"`
	Amount *float64 `json:"amount"`
}

type AssetLogsDelete struct {
	ToAsset     string `json:"to_asset" binding:"required"`
	FromAsset   string `json:"from_asset" binding:"required"`
	AssetMarket string `json:"asset_market" binding:"required"`
}
