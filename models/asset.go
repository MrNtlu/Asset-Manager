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
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AssetModel struct {
	Collection *mongo.Collection
}

func NewAssetModel(mongoDB *db.MongoDB) *AssetModel {
	return &AssetModel{
		Collection: mongoDB.Database.Collection("assets"),
	}
}

type Asset struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	ToAsset       string             `bson:"to_asset" json:"to_asset"`
	FromAsset     string             `bson:"from_asset" json:"from_asset"`
	Price         float64            `bson:"price" json:"price"`
	Amount        float64            `bson:"amount" json:"amount"`
	AssetType     string             `bson:"asset_type" json:"asset_type"`
	AssetMarket   string             `bson:"asset_market" json:"asset_market"`
	Type          string             `bson:"type" json:"type"`
	CurrencyValue float64            `bson:"value" json:"value"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

const (
	assetLogPaginationLimit = 15
	assetPremiumLimit       = 10
)

func createAssetObject(
	uid, toAsset, fromAsset, assetType, assetMarket,
	tType string, price, amount, currencyValue float64,
) *Asset {
	return &Asset{
		UserID:        uid,
		ToAsset:       toAsset,
		FromAsset:     fromAsset,
		Price:         price,
		Amount:        amount,
		AssetType:     assetType,
		AssetMarket:   assetMarket,
		Type:          tType,
		CreatedAt:     time.Now().UTC(),
		CurrencyValue: currencyValue,
	}
}

func (assetModel *AssetModel) CreateAsset(uid string, data requests.AssetCreate) error {
	currencyValue := data.Price * data.Amount

	asset := createAssetObject(
		uid,
		data.ToAsset,
		strings.ToUpper(data.FromAsset),
		data.AssetType,
		data.AssetMarket,
		data.Type,
		data.Price,
		data.Amount,
		currencyValue,
	)

	if _, err := assetModel.Collection.InsertOne(context.TODO(), asset); err != nil {
		logrus.WithFields(logrus.Fields{
			"asset": asset,
		}).Error("failed to create new asset: ", err)

		return fmt.Errorf("Failed to create new asset.")
	}

	return nil
}

func (assetModel *AssetModel) GetAssetByID(assetID string) (Asset, error) {
	objectAssetID, _ := primitive.ObjectIDFromHex(assetID)

	result := assetModel.Collection.FindOne(context.TODO(), bson.M{"_id": objectAssetID})

	var asset Asset
	if err := result.Decode(&asset); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": assetID,
		}).Error("failed to find asset by asset id: ", err)

		return Asset{}, fmt.Errorf("Failed to find asset by asset id.")
	}

	return asset, nil
}

func (assetModel *AssetModel) GetUserAssetCount(uid string) int64 {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}
	facet := bson.M{"$facet": bson.M{
		"document_count": bson.A{
			bson.M{
				"$group": bson.M{
					"_id": bson.M{
						"to_asset":   "$to_asset",
						"from_asset": "$from_asset",
					},
				},
			},
			bson.M{
				"$group": bson.M{
					"_id": nil,
					"n": bson.M{
						"$sum": 1,
					},
				},
			},
		},
	}}

	cursor, err := assetModel.Collection.Aggregate(context.TODO(), bson.A{match, facet})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate assets while counting: ", err)

		return assetPremiumLimit
	}

	var dcArray []responses.AssetDocumentCount
	if err = cursor.All(context.TODO(), &dcArray); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode assets while counting: ", err)

		return assetPremiumLimit
	}

	if len(dcArray) > 0 && len(dcArray[0].DocumentCount) > 0 {
		return dcArray[0].DocumentCount[0].N
	}

	return 0
}

func (assetModel *AssetModel) GetAssetsByUserID(uid string, data requests.AssetSortFilter) ([]responses.Asset, error) {
	var sort bson.M

	switch data.Sort {
	case "name":
		sort = bson.M{"$sort": bson.M{
			"to_asset": data.SortType,
		}}
	case "amount":
		sort = bson.M{"$sort": bson.M{
			"remaining_amount": data.SortType,
		}}
	case "percentage":
		sort = bson.M{"$sort": bson.M{
			"remaining_amount": data.SortType,
		}}
	case "profit":
		sort = bson.M{"$sort": bson.M{
			"p/l": data.SortType,
		}}
	}

	var match bson.M

	if data.AssetType != nil {
		assetTypeList := bson.A{}
		for _, item := range strings.Split(*data.AssetType, ",") {
			assetTypeList = append(assetTypeList, item)
		}

		match = bson.M{"$match": bson.M{
			"user_id": uid,
			"asset_type": bson.M{
				"$in": assetTypeList,
			},
		}}
	} else {
		match = bson.M{"$match": bson.M{
			"user_id": uid,
		}}
	}

	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"to_asset":   "$to_asset",
			"from_asset": "$from_asset",
		},
		"total_bought": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$type", "buy"}},
					"$value",
					0,
				},
			},
		},
		"total_sold": bson.M{
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
		"asset_market": bson.M{
			"$first": "$asset_market",
		},
	}}
	lookup := bson.M{"$lookup": bson.M{
		"from": "investings",
		"let": bson.M{
			"to_asset":   "$_id.to_asset",
			"asset_type": "$asset_type",
			"market":     "$asset_market",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$_id.symbol", "$$to_asset"}},
							bson.M{"$eq": bson.A{"$_id.type", "$$asset_type"}},
							bson.M{"$eq": bson.A{"$_id.market", "$$market"}},
						},
					},
				},
			},
		},
		"as": "investing",
	}}
	unwindInvesting := bson.M{"$unwind": bson.M{
		"path":                       "$investing",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	exchangeLookup := bson.M{"$lookup": bson.M{
		"from": "exchanges",
		"let": bson.M{
			"to_asset":   "$_id.to_asset",
			"from_asset": "$_id.from_asset",
			"asset_type": "$asset_type",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$cond": bson.A{
							bson.M{"$eq": bson.A{"$$asset_type", "exchange"}},
							bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$from_exchange", "$$to_asset"}},
									bson.M{"$eq": bson.A{"$to_exchange", "$$from_asset"}},
								},
							},
							nil,
						},
					},
				},
			},
		},
		"as": "investing_exchange",
	}}
	unwindExchange := bson.M{"$unwind": bson.M{
		"path":                       "$investing_exchange",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	addInvestingField := bson.M{"$addFields": bson.M{
		"investing_price": bson.M{
			"$cond": bson.A{
				bson.M{
					"$eq": bson.A{"$asset_type", "exchange"},
				},
				bson.M{
					"$multiply": bson.A{
						"$investing.price",
						"$investing_exchange.exchange_rate",
					},
				},
				"$investing.price",
			},
		},
		"remaining_amount": bson.M{
			"$cond": bson.A{
				bson.M{
					"$gt": bson.A{"$remaining_amount", 0},
				},
				"$remaining_amount",
				0,
			},
		},
	}}
	project := bson.M{"$project": bson.M{
		"to_asset":         "$_id.to_asset",
		"from_asset":       "$_id.from_asset",
		"name":             "$investing.name",
		"asset_type":       true,
		"asset_market":     true,
		"total_bought":     true,
		"total_sold":       true,
		"remaining_amount": true,
		"current_total_value": bson.M{
			"$multiply": bson.A{"$remaining_amount", "$investing_price"},
		},
		"p/l": bson.M{
			"$subtract": bson.A{
				"$total_bought",
				bson.M{
					"$sum": bson.A{
						"$total_sold",
						bson.M{
							"$multiply": bson.A{"$remaining_amount", "$investing_price"},
						},
					},
				},
			},
		},
	}}
	addPercentageField := bson.M{"$addFields": bson.M{
		"pl_percentage": bson.M{
			"$multiply": bson.A{
				bson.M{
					"$divide": bson.A{"$p/l", "$total_bought"},
				},
				100,
			},
		},
	}}

	cursor, err := assetModel.Collection.Aggregate(context.TODO(), bson.A{
		match, group, lookup, unwindInvesting, exchangeLookup,
		unwindExchange, addInvestingField, project, addPercentageField, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":       uid,
			"sort":      data.Sort,
			"sort_type": data.SortType,
		}).Error("failed to aggregate assets: ", err)

		return nil, fmt.Errorf("Failed to aggregate assets.")
	}

	var assets []responses.Asset
	if err = cursor.All(context.TODO(), &assets); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":       uid,
			"sort":      data.Sort,
			"sort_type": data.SortType,
		}).Error("failed to decode assets: ", err)

		return nil, fmt.Errorf("Failed to decode assets.")
	}

	return assets, nil
}

func (assetModel *AssetModel) GetAssetStatsByAssetAndUserID(uid, toAsset, fromAsset, market string) (responses.AssetDetails, error) {
	match := bson.M{"$match": bson.M{
		"to_asset":     toAsset,
		"from_asset":   fromAsset,
		"asset_market": market,
		"user_id":      uid,
	}}
	lookup := bson.M{"$lookup": bson.M{
		"from": "investings",
		"let": bson.M{
			"to_asset":   "$_id.to_asset",
			"asset_type": "$asset_type",
			"market":     "$asset_market",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$_id.symbol", "$$to_asset"}},
							bson.M{"$eq": bson.A{"$_id.type", "$$asset_type"}},
							bson.M{"$eq": bson.A{"$_id.market", "$$market"}},
						},
					},
				},
			},
		},
		"as": "investing",
	}}
	unwindInvesting := bson.M{"$unwind": bson.M{
		"path":                       "$investing",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	exchangeLookup := bson.M{"$lookup": bson.M{
		"from": "exchanges",
		"let": bson.M{
			"to_asset":   "$_id.to_asset",
			"from_asset": "$_id.from_asset",
			"asset_type": "$asset_type",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$cond": bson.A{
							bson.M{"$eq": bson.A{"$$asset_type", "exchange"}},
							bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$from_exchange", "$$to_asset"}},
									bson.M{"$eq": bson.A{"$to_exchange", "$$from_asset"}},
								},
							},
							nil,
						},
					},
				},
			},
		},
		"as": "investing_exchange",
	}}
	unwindExchange := bson.M{"$unwind": bson.M{
		"path":                       "$investing_exchange",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	addInvestingField := bson.M{"$addFields": bson.M{
		"investing_price": bson.M{
			"$cond": bson.A{
				bson.M{
					"$eq": bson.A{"$asset_type", "exchange"},
				},
				bson.M{
					"$multiply": bson.A{
						"$investing.price",
						"$investing_exchange.exchange_rate",
					},
				},
				"$investing.price",
			},
		},
		"remaining_amount": bson.M{
			"$cond": bson.A{
				bson.M{
					"$gt": bson.A{"$remaining_amount", 0},
				},
				"$remaining_amount",
				0,
			},
		},
	}}
	project := bson.M{"$project": bson.M{
		"to_asset":         "$_id.to_asset",
		"from_asset":       "$_id.from_asset",
		"name":             "$investing.name",
		"total_bought":     true,
		"total_sold":       true,
		"remaining_amount": true,
		"asset_type":       true,
		"asset_market":     true,
		"current_total_value": bson.M{
			"$multiply": bson.A{"$remaining_amount", "$investing_price"},
		},
		"p/l": bson.M{
			"$subtract": bson.A{
				"$total_bought",
				bson.M{
					"$sum": bson.A{
						"$total_sold",
						bson.M{
							"$multiply": bson.A{"$remaining_amount", "$investing_price"},
						},
					},
				},
			},
		},
	}}
	addPercentageField := bson.M{"$addFields": bson.M{
		"pl_percentage": bson.M{
			"$multiply": bson.A{
				bson.M{
					"$divide": bson.A{"$p/l", "$total_bought"},
				},
				100,
			},
		},
	}}

	cursor, err := assetModel.Collection.Aggregate(context.TODO(), bson.A{
		match, groupAssetsByToAssetFromAsset(), lookup, unwindInvesting, exchangeLookup,
		unwindExchange, addInvestingField, project, addPercentageField,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":        uid,
			"to_asset":   toAsset,
			"from_asset": fromAsset,
		}).Error("failed to aggregate asset details: ", err)

		return responses.AssetDetails{}, fmt.Errorf("Failed to aggregate asset details.")
	}

	var assetDetails []responses.AssetDetails
	if err = cursor.All(context.TODO(), &assetDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":        uid,
			"to_asset":   toAsset,
			"from_asset": fromAsset,
		}).Error("failed to decode asset details: ", err)

		return responses.AssetDetails{}, fmt.Errorf("Failed to decode asset details.")
	}

	if len(assetDetails) > 0 {
		return assetDetails[0], nil
	}

	return responses.AssetDetails{}, nil
}

func (assetModel *AssetModel) GetAllAssetStats(uid string) (responses.AssetStats, error) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}
	lookup := bson.M{"$lookup": bson.M{
		"from": "investings",
		"let": bson.M{
			"to_asset":   "$_id.to_asset",
			"asset_type": "$asset_type",
			"market":     "$asset_market",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$_id.symbol", "$$to_asset"}},
							bson.M{"$eq": bson.A{"$_id.type", "$$asset_type"}},
							bson.M{"$eq": bson.A{"$_id.market", "$$market"}},
						},
					},
				},
			},
		},
		"as": "investing",
	}}
	unwindInvesting := bson.M{"$unwind": bson.M{
		"path":                       "$investing",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	exchangeLookup := bson.M{"$lookup": bson.M{
		"from": "exchanges",
		"let": bson.M{
			"to_asset":   "$_id.to_asset",
			"from_asset": "$_id.from_asset",
			"asset_type": "$asset_type",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$cond": bson.A{
							bson.M{"$eq": bson.A{"$$asset_type", "exchange"}},
							bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$from_exchange", "$$to_asset"}},
									bson.M{"$eq": bson.A{"$to_exchange", "$$from_asset"}},
								},
							},
							nil,
						},
					},
				},
			},
		},
		"as": "investing_exchange",
	}}
	unwindExchange := bson.M{"$unwind": bson.M{
		"path":                       "$investing_exchange",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	addInvestingField := bson.M{"$addFields": bson.M{
		"investing_price": bson.M{
			"$cond": bson.A{
				bson.M{
					"$eq": bson.A{"$asset_type", "exchange"},
				},
				bson.M{
					"$multiply": bson.A{
						"$investing.price",
						"$investing_exchange.exchange_rate",
					},
				},
				"$investing.price",
			},
		},
		"remaining_amount": bson.M{
			"$cond": bson.A{
				bson.M{
					"$gt": bson.A{"$remaining_amount", 0},
				},
				"$remaining_amount",
				0,
			},
		},
	}}
	project := bson.M{"$project": bson.M{
		"user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"total_bought": true,
		"total_sold":   true,
		"asset_type":   true,
		"current_total_value": bson.M{
			"$multiply": bson.A{"$remaining_amount", "$investing_price"},
		},
		"p/l": bson.M{
			"$subtract": bson.A{
				"$total_bought",
				bson.M{
					"$sum": bson.A{
						"$total_sold",
						bson.M{
							"$multiply": bson.A{"$remaining_amount", "$investing_price"},
						},
					},
				},
			},
		},
	}}
	userLookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "user_id",
		"foreignField": "_id",
		"as":           "user",
	}}
	unwindUser := bson.M{"$unwind": bson.M{
		"path":                       "$user",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	userCurrencyExchangeLookup := bson.M{"$lookup": bson.M{
		"from": "exchanges",
		"let": bson.M{
			"user_currency": "$user.currency",
			"from_asset":    "$_id.from_asset",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$ne": bson.A{"$$from_asset", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$to_exchange", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$from_exchange", "$$from_asset"}},
						},
					},
				},
			},
		},
		"as": "user_exchange_rate",
	}}
	unwindUserCurrency := bson.M{"$unwind": bson.M{
		"path":                       "$user_exchange_rate",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	userCurrencyProject := bson.M{"$project": bson.M{
		"user_id":    true,
		"asset_type": true,
		"currency":   "$user.currency",
		"total_assets": bson.M{
			"$ifNull": bson.A{
				bson.M{
					"$multiply": bson.A{"$current_total_value", "$user_exchange_rate.exchange_rate"},
				},
				"$current_total_value",
			},
		},
		"total_p/l": bson.M{
			"$ifNull": bson.A{
				bson.M{
					"$multiply": bson.A{"$p/l", "$user_exchange_rate.exchange_rate"},
				},
				"$p/l",
			},
		},
		"total_bought": bson.M{
			"$ifNull": bson.A{
				bson.M{
					"$multiply": bson.A{"$total_bought", "$user_exchange_rate.exchange_rate"},
				},
				"$total_bought",
			},
		},
		"total_sold": bson.M{
			"$ifNull": bson.A{
				bson.M{
					"$multiply": bson.A{"$total_sold", "$user_exchange_rate.exchange_rate"},
				},
				"$total_sold",
			},
		},
	}}
	assetGroup := bson.M{"$group": bson.M{
		"_id": "$asset_type",
		"currency": bson.M{
			"$first": "$currency",
		},
		"total_bought": bson.M{
			"$sum": "$total_bought",
		},
		"total_sold": bson.M{
			"$sum": "$total_sold",
		},
		"total_assets": bson.M{
			"$sum": "$total_assets",
		},
		"total_p/l": bson.M{
			"$sum": "$total_p/l",
		},
		"user_id": bson.M{
			"$first": "$user_id",
		},
	}}
	statsGroup := bson.M{"$group": bson.M{
		"_id": "$user_id",
		"currency": bson.M{
			"$first": "$currency",
		},
		"total_bought": bson.M{
			"$sum": "$total_bought",
		},
		"total_sold": bson.M{
			"$sum": "$total_sold",
		},
		"stock_assets": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$_id", "stock"}},
					"$total_assets",
					0,
				},
			},
		},
		"crypto_assets": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$_id", "crypto"}},
					"$total_assets",
					0,
				},
			},
		},
		"exchange_assets": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$_id", "exchange"}},
					"$total_assets",
					0,
				},
			},
		},
		"commodity_assets": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$_id", "commodity"}},
					"$total_assets",
					0,
				},
			},
		},
		"total_assets": bson.M{
			"$sum": "$total_assets",
		},
		"stock_p/l": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$_id", "stock"}},
					"$total_p/l",
					0,
				},
			},
		},
		"crypto_p/l": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$_id", "crypto"}},
					"$total_p/l",
					0,
				},
			},
		},
		"exchange_p/l": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$_id", "exchange"}},
					"$total_p/l",
					0,
				},
			},
		},
		"commodity_p/l": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$_id", "commodity"}},
					"$total_p/l",
					0,
				},
			},
		},
		"total_p/l": bson.M{
			"$sum": "$total_p/l",
		},
	}}
	addPercentageFields := bson.M{"$addFields": bson.M{
		"total_pl_percentage": bson.M{
			"$multiply": bson.A{
				bson.M{
					"$cond": bson.A{
						bson.M{"$ne": bson.A{"$total_bought", 0}},
						bson.M{"$divide": bson.A{"$total_p/l", "$total_bought"}},
						0,
					},
				},
				100,
			},
		},
		"stock_percentage": bson.M{
			"$multiply": bson.A{
				bson.M{
					"$cond": bson.A{
						bson.M{"$ne": bson.A{"$total_assets", 0}},
						bson.M{"$divide": bson.A{
							"$stock_assets", "$total_assets",
						}},
						0,
					},
				},
				100,
			},
		},
		"crypto_percentage": bson.M{
			"$multiply": bson.A{
				bson.M{
					"$cond": bson.A{
						bson.M{"$ne": bson.A{"$total_assets", 0}},
						bson.M{"$divide": bson.A{
							"$crypto_assets", "$total_assets",
						}},
						0,
					},
				},
				100,
			},
		},
		"exchange_percentage": bson.M{
			"$multiply": bson.A{
				bson.M{
					"$cond": bson.A{
						bson.M{"$ne": bson.A{"$total_assets", 0}},
						bson.M{"$divide": bson.A{
							"$exchange_assets", "$total_assets",
						}},
						0,
					},
				},
				100,
			},
		},
		"commodity_percentage": bson.M{
			"$multiply": bson.A{
				bson.M{
					"$cond": bson.A{
						bson.M{"$ne": bson.A{"$total_assets", 0}},
						bson.M{"$divide": bson.A{
							"$commodity_assets", "$total_assets",
						}},
						0,
					},
				},
				100,
			},
		},
	}}

	cursor, err := assetModel.Collection.Aggregate(context.TODO(), bson.A{
		match, groupAssetsByToAssetFromAsset(), lookup, unwindInvesting, exchangeLookup, unwindExchange,
		addInvestingField, project, userLookup, unwindUser, userCurrencyExchangeLookup,
		unwindUserCurrency, userCurrencyProject, assetGroup, statsGroup, addPercentageFields,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate asset stats: ", err)

		return responses.AssetStats{}, fmt.Errorf("Failed to aggregate asset stats.")
	}

	var assetStat []responses.AssetStats
	if err = cursor.All(context.TODO(), &assetStat); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode asset stats: ", err)

		return responses.AssetStats{}, fmt.Errorf("Failed to decode asset stats.")
	}

	if len(assetStat) > 0 {
		return assetStat[0], nil
	}

	return responses.AssetStats{}, nil
}

func (assetModel *AssetModel) GetAssetLogsByUserID(uid string, data requests.AssetLog) ([]Asset, pagination.PaginationData, error) {
	match := bson.M{
		"to_asset":     data.ToAsset,
		"from_asset":   data.FromAsset,
		"asset_market": data.AssetMarket,
		"user_id":      uid,
	}

	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "newest":
		sortType = "created_at"
		sortOrder = -1
	case "oldest":
		sortType = "created_at"
		sortOrder = 1
	case "amount":
		sortType = "amount"
		sortOrder = -1
	}

	var assets []Asset

	paginatedData, err := pagination.New(assetModel.Collection).Context(context.TODO()).
		Limit(assetLogPaginationLimit).Sort(sortType, sortOrder).Page(data.Page).Filter(match).Decode(&assets).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":        uid,
			"to_asset":   data.ToAsset,
			"from_asset": data.FromAsset,
			"page":       data.Page,
			"sort":       data.Sort,
		}).Error("failed to fetch/decode: ", err)

		return nil, pagination.PaginationData{}, fmt.Errorf("Failed to get asset logs.")
	}

	return assets, paginatedData.Pagination, nil
}

func (assetModel *AssetModel) UpdateAssetLogByAssetID(data requests.AssetUpdate, asset Asset) error {
	objectAssetID, _ := primitive.ObjectIDFromHex(data.ID)

	if data.Type != nil {
		asset.Type = *data.Type
	}

	if data.Price != nil {
		asset.Price = *data.Price
	}

	if data.Amount != nil {
		asset.Amount = *data.Amount
	}

	if data.Amount != nil || data.Price != nil {
		asset.CurrencyValue = asset.Price * asset.Amount
	}

	if _, err := assetModel.Collection.UpdateOne(context.TODO(), bson.M{
		"_id": objectAssetID,
	}, bson.M{"$set": asset}); err != nil {
		logrus.WithFields(logrus.Fields{
			"asset_id": data.ID,
			"data":     data,
		}).Error("failed to update asset: ", err)

		return fmt.Errorf("Failed to update asset.")
	}

	return nil
}

func (assetModel *AssetModel) DeleteAssetLogByAssetID(uid, assetID string) (bool, error) {
	objectAssetID, _ := primitive.ObjectIDFromHex(assetID)

	count, err := assetModel.Collection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectAssetID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":      uid,
			"asset_id": assetID,
		}).Error("failed to delete asset: ", err)

		return false, fmt.Errorf("Failed to delete asset.")
	}

	return count.DeletedCount > 0, nil
}

func (assetModel *AssetModel) DeleteAssetLogsByUserID(uid string, data requests.AssetLogsDelete) error {
	if _, err := assetModel.Collection.DeleteMany(context.TODO(), bson.M{
		"to_asset":     data.ToAsset,
		"from_asset":   data.FromAsset,
		"asset_market": data.AssetMarket,
		"user_id":      uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":        uid,
			"to_asset":   data.ToAsset,
			"from_asset": data.FromAsset,
		}).Error("failed to delete asset logs by user id: ", err)

		return fmt.Errorf("Failed to delete asset logs by user.")
	}

	return nil
}

func (assetModel *AssetModel) DeleteAllAssetsByUserID(uid string) error {
	if _, err := assetModel.Collection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all assets by user id: ", err)

		return fmt.Errorf("Failed to delete all assets by user.")
	}

	return nil
}

func groupAssetsByToAssetFromAsset() bson.M {
	return bson.M{"$group": bson.M{
		"_id": bson.M{
			"to_asset":   "$to_asset",
			"from_asset": "$from_asset",
		},
		"total_bought": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$type", "buy"}},
					"$value",
					0,
				},
			},
		},
		"total_sold": bson.M{
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
		"asset_market": bson.M{
			"$first": "$asset_market",
		},
		"user_id": bson.M{
			"$first": "$user_id",
		},
	}}
}
