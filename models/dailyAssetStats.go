package models

import (
	"asset_backend/db"
	"asset_backend/responses"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAssetStatsByUserID(uid string, interval string) (responses.DailyAssetStats, error) {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	var intervalDate time.Time

	switch interval {
	case "weekly":
		intervalDate = time.Now().AddDate(0, 0, -7)
	case "monthly":
		intervalDate = time.Now().AddDate(0, -1, 0)
	}

	var match bson.M
	if interval == "yearly" {
		match = bson.M{"$match": bson.M{
			"user_id": objectUID,
			"$expr": bson.M{
				"$eq": bson.A{
					bson.M{"$year": "$created_at"},
					time.Now().Year(),
				},
			},
		}}
	} else {
		match = bson.M{"$match": bson.M{
			"user_id": objectUID,
			"created_at": bson.M{
				"$gte": intervalDate,
			},
		}}
	}

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
	exchangeLookup := bson.M{"$lookup": bson.M{
		"from": "exchanges",
		"let": bson.M{
			"user_currency": "$user.currency",
			"stat_currency": "$currency",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$ne": bson.A{"$$stat_currency", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$to_exchange", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$from_exchange", "$$stat_currency"}},
						},
					},
				},
			},
		},
		"as": "user_exchange_rate",
	}}
	unwindExchange := bson.M{"$unwind": bson.M{
		"path":                       "$user_exchange_rate",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	var (
		project     bson.M
		yearlyGroup bson.M
	)

	if interval == "yearly" {
		project = bson.M{"$project": bson.M{
			"currency": "$user.currency",
			"total_assets": bson.M{
				"$ifNull": bson.A{
					bson.M{
						"$multiply": bson.A{"$total_assets", "$user_exchange_rate.exchange_rate"},
					},
					"$total_assets",
				},
			},
			"total_p/l": bson.M{
				"$ifNull": bson.A{
					bson.M{
						"$multiply": bson.A{"$total_p/l", "$user_exchange_rate.exchange_rate"},
					},
					"$total_p/l",
				},
			},
			"created_at": bson.M{
				"$dateTrunc": bson.M{
					"date": "$created_at",
					"unit": "month",
				},
			},
		}}
		yearlyGroup = bson.M{"$group": bson.M{
			"_id": "$created_at",
			"currency": bson.M{
				"$first": "$currency",
			},
			"total_p/l": bson.M{
				"$sum": "$total_p/l",
			},
			"total_assets": bson.M{
				"$sum": "$total_assets",
			},
			"created_at": bson.M{
				"$first": "$created_at",
			},
		}}
	} else {
		project = bson.M{"$project": bson.M{
			"currency":   "$user.currency",
			"created_at": true,
			"total_assets": bson.M{
				"$ifNull": bson.A{
					bson.M{
						"$multiply": bson.A{"$total_assets", "$user_exchange_rate.exchange_rate"},
					},
					"$total_assets",
				},
			},
			"total_p/l": bson.M{
				"$ifNull": bson.A{
					bson.M{
						"$multiply": bson.A{"$total_p/l", "$user_exchange_rate.exchange_rate"},
					},
					"$total_p/l",
				},
			},
		}}
	}

	sort := bson.M{"$sort": bson.M{
		"created_at": 1,
	}}
	arrayGroup := bson.M{"$group": bson.M{
		"_id": nil,
		"currency": bson.M{
			"$first": "$currency",
		},
		"dates": bson.M{
			"$push": "$created_at",
		},
		"total_assets": bson.M{
			"$push": "$total_assets",
		},
		"total_p/l": bson.M{
			"$push": "$total_p/l",
		},
	}}

	var aggregationList bson.A
	if interval == "yearly" {
		aggregationList = bson.A{
			match, userLookup, unwindUser, exchangeLookup, unwindExchange, project, yearlyGroup, sort, arrayGroup,
		}
	} else {
		aggregationList = bson.A{
			match, userLookup, unwindUser, exchangeLookup, unwindExchange, project, sort, arrayGroup,
		}
	}

	cursor, err := db.DailyAssetStatCollection.Aggregate(context.TODO(), aggregationList)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":      uid,
			"interval": interval,
		}).Error("failed to aggregate daily asset stats: ", err)

		return responses.DailyAssetStats{}, fmt.Errorf("Failed to aggregate daily asset stats.")
	}

	var dailyAssetStats []responses.DailyAssetStats
	if err = cursor.All(context.TODO(), &dailyAssetStats); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":      uid,
			"interval": interval,
		}).Error("failed to decode daily asset stats: ", err)

		return responses.DailyAssetStats{}, fmt.Errorf("Failed to decode daily asset stats.")
	}

	if len(dailyAssetStats) > 0 {
		return dailyAssetStats[0], nil
	}

	return responses.DailyAssetStats{}, nil
}

func CalculateDailyAssetStats() {
	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"to_asset":   "$to_asset",
			"from_asset": "$from_asset",
			"user_id":    "$user_id",
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
		"user_id":      true,
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
		"user_id": bson.M{
			"$toString": "$user_id",
		},
		"asset_type": true,
		"created_at": time.Now().UTC(),
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
	}}
	assetGroup := bson.M{"$group": bson.M{
		"_id": bson.M{
			"asset_type": "$asset_type",
			"user_id":    "$user_id",
		},
		"currency": bson.M{
			"$first": "$currency",
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
		"created_at": bson.M{
			"$first": "$created_at",
		},
	}}
	statsGroup := bson.M{"$group": bson.M{
		"_id": "$user_id",
		"currency": bson.M{
			"$first": "$currency",
		},
		"total_assets": bson.M{
			"$sum": "$total_assets",
		},
		"total_p/l": bson.M{
			"$sum": "$total_p/l",
		},
		"created_at": bson.M{
			"$first": "$created_at",
		},
		"user_id": bson.M{
			"$first": "$user_id",
		},
	}}

	cursor, err := db.AssetCollection.Aggregate(context.TODO(), bson.A{
		group, lookup, unwindInvesting, exchangeLookup, unwindExchange,
		addInvestingField, project, userLookup, unwindUser, userCurrencyExchangeLookup,
		unwindUserCurrency, userCurrencyProject, assetGroup, statsGroup,
	})
	if err != nil {
		logrus.Error("failed to aggregate daily asset stats calculation: ", err)
	}

	var dailyAssetStats []responses.DailyAssetStatsCalculation
	if err = cursor.All(context.TODO(), &dailyAssetStats); err != nil {
		logrus.Error("failed to decode daily asset stats calculation: ", err)
	}

	if len(dailyAssetStats) < 1 {
		logrus.Error("empty daily asset stats")
		return
	}

	insertDASList := make([]interface{}, len(dailyAssetStats))
	for i, dailyAssetStat := range dailyAssetStats {
		insertDASList[i] = dailyAssetStat
	}

	if _, err := db.DailyAssetStatCollection.InsertMany(
		context.TODO(),
		insertDASList,
		options.InsertMany().SetOrdered(false),
	); err != nil {
		logrus.Error("failed to create daily asset stats calculation list: ", err)
	}
}

func DeleteAllAssetStatsByUserID(uid string) error {
	if _, err := db.DailyAssetStatCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all asset stats by user id: ", err)

		return fmt.Errorf("Failed to delete all asset stats by user id.")
	}

	return nil
}
