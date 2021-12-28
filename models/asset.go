package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//var assetCollection = db.Database.Collection("assets")

type Asset struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	ToAsset     string             `bson:"to_asset" json:"to_asset"`
	FromAsset   string             `bson:"from_asset" json:"from_asset"`
	BoughtPrice float64            `bson:"bought_price" json:"bought_price"`
	SoldPrice   *float64           `bson:"sold_price" json:"sold_price"`
	Amount      float64            `bson:"amount" json:"amount"`
	AssetType   string             `bson:"asset_type" json:"asset_type"`
	Type        string             `bson:"type" json:"type"`
	CreatedAt   time.Time          `bson:"created_at" json:"-"`
}

func createAssetObject(uid, toAsset, fromAsset, assetType, tType string, boughtPrice, amount float64, soldPrice *float64) *Asset {
	return &Asset{
		UserID:      uid,
		ToAsset:     toAsset,
		FromAsset:   fromAsset,
		BoughtPrice: boughtPrice,
		SoldPrice:   soldPrice,
		Amount:      amount,
		AssetType:   assetType,
		Type:        tType,
		CreatedAt:   time.Now().UTC(),
	}
}

func CreateAsset(data requests.Asset, uid string) error {
	asset := createAssetObject(
		uid,
		data.ToAsset,
		data.FromAsset,
		data.AssetType,
		data.Type,
		data.BoughtPrice,
		data.Amount,
		data.SoldPrice,
	)

	if _, err := db.AssetCollection.InsertOne(context.TODO(), asset); err != nil {
		return fmt.Errorf("failed to create new asset: %w", err)
	}

	return nil
}

func GetAssetsByUserID(uid string) ([]responses.Asset, error) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}
	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"to_asset":   "$to_asset",
			"from_asset": "$from_asset",
		},
		"amount": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$type", "buy"}},
					"$amount",
					bson.M{"$multiply": bson.A{"$amount", -1}},
				},
			},
		},
		"asset_type": bson.M{
			"$first": "$asset_type",
		},
	}}
	project := bson.M{"$project": bson.M{
		"to_asset":   "$_id.to_asset",
		"from_asset": "$_id.from_asset",
		"amount":     true,
		"asset_type": true,
	}}

	cursor, err := db.AssetCollection.Aggregate(context.TODO(), bson.A{match, group, project})
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate assets: %w", err)
	}

	var assets []responses.Asset
	if err = cursor.All(context.TODO(), &assets); err != nil {
		return nil, fmt.Errorf("failed to decode asset: %w", err)
	}

	return assets, nil
}

func GetAssetLogsByID(assetID string) ([]Asset, error) { //TODO: Pagination
	objectAssetID, _ := primitive.ObjectIDFromHex(assetID)

	cursor, err := db.AssetCollection.Find(context.TODO(), bson.M{"_id": objectAssetID})
	if err != nil {
		return nil, fmt.Errorf("failed to find asset: %w", err)
	}

	var assets []Asset
	if err := cursor.All(context.TODO(), &assets); err != nil {
		return nil, fmt.Errorf("failed to decode asset: %w", err)
	}

	return assets, nil
}
