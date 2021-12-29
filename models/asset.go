package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"fmt"
	"strings"
	"time"

	pagination "github.com/gobeam/mongo-go-pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//var assetCollection = db.Database.Collection("assets")

type Asset struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	ToAsset     string             `bson:"to_asset" json:"to_asset"`
	FromAsset   string             `bson:"from_asset" json:"from_asset"`
	BoughtPrice *float64           `bson:"bought_price" json:"bought_price"`
	SoldPrice   *float64           `bson:"sold_price" json:"sold_price"`
	Amount      float64            `bson:"amount" json:"amount"`
	AssetType   string             `bson:"asset_type" json:"asset_type"`
	Type        string             `bson:"type" json:"type"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

func createAssetObject(uid, toAsset, fromAsset, assetType, tType string, amount float64, boughtPrice, soldPrice *float64) *Asset {
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

func CreateAsset(data requests.AssetCreate, uid string) error {
	asset := createAssetObject(
		uid,
		strings.ToUpper(data.ToAsset),
		strings.ToUpper(data.FromAsset),
		data.AssetType,
		data.Type,
		data.Amount,
		data.BoughtPrice,
		data.SoldPrice,
	)

	if _, err := db.AssetCollection.InsertOne(context.TODO(), asset); err != nil {
		return fmt.Errorf("failed to create new asset: %w", err)
	}

	return nil
}

func GetAssetsByUserID(uid string, data requests.AssetSort) ([]responses.Asset, error) {
	var sort bson.M
	if data.Sort == "name" {
		sort = bson.M{"$sort": bson.M{
			"to_asset": 1,
		}}
	} else if data.Sort == "amount" {
		sort = bson.M{"$sort": bson.M{
			"amount": -1,
		}}
	} else if data.Sort == "worth" {
		sort = bson.M{"$sort": bson.M{
			"avg_worth": -1,
		}}
	} else {
		sort = bson.M{"$sort": bson.M{
			"avg_worth": 1,
		}}
	}

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
		"avg_price": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$type", "buy"}},
					bson.M{"$multiply": bson.A{
						"$bought_price", "$amount",
					}},
					0,
				},
			},
		},
		"asset_type": bson.M{
			"$first": "$asset_type",
		},
	}}
	project := bson.M{"$project": bson.M{
		"avg_price": bson.M{
			"$round": bson.A{
				bson.M{"$divide": bson.A{"$avg_price", "$amount"}},
				2,
			}},
		"to_asset":   "$_id.to_asset",
		"from_asset": "$_id.from_asset",
		"amount":     true,
		"asset_type": true,
	}}
	addAvgWorthField := bson.M{"$addFields": bson.M{
		"avg_worth": bson.M{
			"$round": bson.A{
				bson.M{"$multiply": bson.A{"$avg_price", "$amount"}},
				2,
			},
		},
	}}

	cursor, err := db.AssetCollection.Aggregate(context.TODO(), bson.A{match, group, project, addAvgWorthField, sort})
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate assets: %w", err)
	}

	var assets []responses.Asset
	if err = cursor.All(context.TODO(), &assets); err != nil {
		return nil, fmt.Errorf("failed to decode asset: %w", err)
	}

	return assets, nil
}

func GetAssetLogsByUserID(uid string, data requests.AssetLog) ([]Asset, pagination.PaginationData, error) {
	var match bson.M
	if data.Type == nil {
		match = bson.M{
			"user_id":    uid,
			"to_asset":   data.ToAsset,
			"from_asset": data.FromAsset,
		}
	} else {
		match = bson.M{
			"user_id":    uid,
			"to_asset":   data.ToAsset,
			"from_asset": data.FromAsset,
			"type":       data.Type,
		}
	}

	var sortType string
	var sortOrder int8
	if data.Sort == "newest" {
		sortType = "created_at"
		sortOrder = -1
	} else if data.Sort == "oldest" {
		sortType = "created_at"
		sortOrder = 1
	} else {
		sortType = "amount"
		sortOrder = -1
	}

	var assets []Asset
	paginatedData, err := pagination.New(db.AssetCollection).Context(context.TODO()).
		Limit(15).Sort(sortType, sortOrder).Page(data.Page).Filter(match).Decode(&assets).Find()
	if err != nil {
		return nil, pagination.PaginationData{}, fmt.Errorf("failed to fetct/decode: %w", err)
	}

	return assets, paginatedData.Pagination, nil
}

func UpdateAssetLogByAssetID(data requests.AssetUpdate) error {
	objectAssetID, _ := primitive.ObjectIDFromHex(data.AssetID)

	var update bson.M
	if data.BoughtPrice != nil && data.Amount != 0 {
		update = bson.M{
			"bought_price": data.BoughtPrice,
			"amount":       data.Amount,
		}
	} else if data.SoldPrice != nil && data.Amount != 0 {
		update = bson.M{
			"sold_price": data.SoldPrice,
			"amount":     data.Amount,
		}
	} else if data.Amount != 0 {
		update = bson.M{
			"amount": data.Amount,
		}
	} else {
		return nil
	}

	if _, err := db.AssetCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectAssetID,
	}, bson.M{"$set": update}); err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	return nil
}

func DeleteAssetLogByAssetID(assetID string) error {
	objectAssetID, _ := primitive.ObjectIDFromHex(assetID)

	if _, err := db.AssetCollection.DeleteOne(context.TODO(), bson.M{"_id": objectAssetID}); err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	return nil
}

func DeleteAssetLogsByUserID(uid string, data requests.AssetLogsDelete) error {
	if _, err := db.AssetCollection.DeleteMany(context.TODO(), bson.M{
		"user_id":    uid,
		"to_asset":   data.ToAsset,
		"from_asset": data.FromAsset,
	}); err != nil {
		return fmt.Errorf("failed to delete asset logs by user id: %w", err)
	}

	return nil
}

func DeleteAllAssetsByUserID(uid string) error {
	if _, err := db.AssetCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		return fmt.Errorf("failed to delete all assets by user id: %w", err)
	}

	return nil
}
