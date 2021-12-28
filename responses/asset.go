package responses

type Asset struct {
	ToAsset   string  `bson:"to_asset" json:"to_asset"`
	FromAsset string  `bson:"from_asset" json:"from_asset"`
	Amount    float64 `bson:"amount" json:"amount"`
	AssetType string  `bson:"asset_type" json:"asset_type"`
}
