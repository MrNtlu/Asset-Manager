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

type Asset struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	ToAsset       string             `bson:"to_asset" json:"to_asset"`
	FromAsset     string             `bson:"from_asset" json:"from_asset"`
	BoughtPrice   *float64           `bson:"bought_price" json:"bought_price"`
	SoldPrice     *float64           `bson:"sold_price" json:"sold_price"`
	Amount        float64            `bson:"amount" json:"amount"`
	AssetType     string             `bson:"asset_type" json:"asset_type"`
	Type          string             `bson:"type" json:"type"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	CurrencyValue float64            `bson:"value" json:"value"`
}

func createAssetObject(uid, toAsset, fromAsset, assetType, tType string, amount, currencyValue float64, boughtPrice, soldPrice *float64) *Asset {
	return &Asset{
		UserID:        uid,
		ToAsset:       toAsset,
		FromAsset:     fromAsset,
		BoughtPrice:   boughtPrice,
		SoldPrice:     soldPrice,
		Amount:        amount,
		AssetType:     assetType,
		Type:          tType,
		CreatedAt:     time.Now().UTC(),
		CurrencyValue: currencyValue,
	}
}

func CreateAsset(uid string, data requests.AssetCreate) error {
	var currencyValue float64
	if data.Type == "buy" {
		currencyValue = *data.BoughtPrice * data.Amount
	} else {
		currencyValue = *data.SoldPrice * data.Amount
	}

	asset := createAssetObject(
		uid,
		strings.ToUpper(data.ToAsset),
		strings.ToUpper(data.FromAsset),
		data.AssetType,
		data.Type,
		data.Amount,
		currencyValue,
		data.BoughtPrice,
		data.SoldPrice,
	)

	if _, err := db.AssetCollection.InsertOne(context.TODO(), asset); err != nil {
		return fmt.Errorf("failed to create new asset: %w", err)
	}

	return nil
}

func GetAssetByID(assetID string) (Asset, error) {
	objectAssetID, _ := primitive.ObjectIDFromHex(assetID)

	result := db.AssetCollection.FindOne(context.TODO(), bson.M{"_id": objectAssetID})

	var asset Asset
	if err := result.Decode(&asset); err != nil {
		return Asset{}, fmt.Errorf("failed to find asset by asset id: %w", err)
	}

	return asset, nil
}

//TODO:if number == 0 or somehow below 0 hide them in mobile/desktop
// amountMatch := bson.M{"$match": bson.M{
// 	"amount": bson.M{
// 		"$gt": 0,
// 	},
// }}
func GetAssetsByUserID(uid string, data requests.AssetSort) ([]responses.Asset, error) {
	var sort bson.M
	if data.Sort == "name" {
		sort = bson.M{"$sort": bson.M{
			"to_asset": data.SortType,
		}}
	} else if data.Sort == "amount" {
		sort = bson.M{"$sort": bson.M{
			"remaining_amount": data.SortType,
		}}
	} else if data.Sort == "value" {
		sort = bson.M{"$sort": bson.M{
			"total_value": data.SortType,
		}}
	} else {
		sort = bson.M{"$sort": bson.M{
			"p/l": data.SortType,
		}}
	}

	js := `function(prices) {
        var sum = 1;
        for (var i = 0; i < prices.length; i++) {
            sum = sum * prices[i];
        }
        return sum;
      }`

	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}
	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"to_asset":   "$to_asset",
			"from_asset": "$from_asset",
		},
		"total_value": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$type", "buy"}},
					"$value",
					0,
				},
			},
		},
		"sold_value": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$type", "sell"}},
					"$value",
					0,
				},
			},
		},
		"remaining_amount": bson.M{
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
	lookup := bson.M{"$lookup": bson.M{
		"from": "investings",
		"let": bson.M{
			"asset_type": "$asset_type",
			"to_asset":   "$_id.to_asset",
			"from_asset": "$_id.from_asset",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$or": bson.A{
							bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$_id.symbol", "$$to_asset"}},
									bson.M{"$eq": bson.A{"$_id.type", "$$asset_type"}},
								},
							},
							bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$_id.symbol", "$$from_asset"}},
									bson.M{"$eq": bson.A{"$_id.type", "exchange"}},
								},
							},
						},
					},
				},
			},
		},
		"as": "investing",
	}}
	addInvestingField := bson.M{"$addFields": bson.M{
		"investing_price": bson.M{
			"$function": bson.M{
				"body": primitive.JavaScript(js),
				"args": bson.A{"$investing.price"},
				"lang": "js",
			},
		},
	}}
	project := bson.M{"$project": bson.M{
		"to_asset":         "$_id.to_asset",
		"from_asset":       "$_id.from_asset",
		"asset_type":       true,
		"total_value":      true,
		"sold_value":       true,
		"remaining_amount": true,
		"current_price":    "$investing_price",
		"p/l": bson.M{
			"$subtract": bson.A{
				"$total_value",
				bson.M{
					"$sum": bson.A{
						"$sold_value",
						bson.M{
							"$multiply": bson.A{"$remaining_amount", "$investing_price"},
						},
					},
				},
			},
		},
	}}

	cursor, err := db.AssetCollection.Aggregate(context.TODO(), bson.A{match, group, lookup, addInvestingField, project, sort})
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate assets: %w", err)
	}

	var assets []responses.Asset
	if err = cursor.All(context.TODO(), &assets); err != nil {
		return nil, fmt.Errorf("failed to decode asset: %w", err)
	}

	return assets, nil
}

//TODO: Total Asset, profit/loss if sold etc. sum of GetAssetsByUserID
// profit/loss = total amount * avg_bought_price - ((sold amount * avg_sold_price) + (remaining amount * current value))
// = total value - (sold value + (remaining amount * current value))
func GetAllAssetStats(uid string) error {

	return nil
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
		return nil, pagination.PaginationData{}, fmt.Errorf("failed to fetch/decode: %w", err)
	}

	return assets, paginatedData.Pagination, nil
}

func UpdateAssetLogByAssetID(data requests.AssetUpdate, asset Asset) error {
	objectAssetID, _ := primitive.ObjectIDFromHex(data.ID)

	var (
		currencyValue float64
		update        bson.M
	)
	if data.BoughtPrice != nil && data.Amount != 0 {
		currencyValue = *data.BoughtPrice * data.Amount

		update = bson.M{
			"bought_price": data.BoughtPrice,
			"amount":       data.Amount,
			"value":        currencyValue,
		}
	} else if data.SoldPrice != nil && data.Amount != 0 {
		currencyValue = *data.SoldPrice * data.Amount

		update = bson.M{
			"sold_price": data.SoldPrice,
			"amount":     data.Amount,
			"value":      currencyValue,
		}
	} else if data.Amount != 0 {
		if asset.Type == "buy" {
			currencyValue = *asset.BoughtPrice * data.Amount
		} else {
			currencyValue = *asset.SoldPrice * data.Amount
		}

		update = bson.M{
			"amount": data.Amount,
			"value":  currencyValue,
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
