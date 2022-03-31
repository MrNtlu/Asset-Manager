package models

import (
	"asset_backend/db"
	"asset_backend/responses"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetInvestingsByTypeAndMarket(tType, market string) ([]responses.InvestingResponse, error) {
	match := bson.M{"$match": bson.M{
		"_id.type":   tType,
		"_id.market": market,
	}}
	project := bson.M{"$project": bson.M{
		"name":   "$name",
		"symbol": "$_id.symbol",
	}}

	cursor, err := db.InvestingCollection.Aggregate(context.TODO(), bson.A{match, project})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": tType,
		}).Error("failed to aggregate investings: %w", err)
		return nil, fmt.Errorf("failed to fetch investings")
	}

	var investings []responses.InvestingResponse
	if err = cursor.All(context.TODO(), &investings); err != nil {
		logrus.WithFields(logrus.Fields{
			"type": tType,
		}).Error("failed to decode investings: %w", err)
		return nil, fmt.Errorf("failed to decode investings")
	}

	return investings, nil
}

func GetInvestingPriceTableByTypeAndMarket(tType, market string) ([]responses.InvestingTableResponse, error) {
	var cursor *mongo.Cursor
	var err error
	if tType == "exchange" {
		project := bson.M{"$project": bson.M{
			"name":     "$from_exchange",
			"symbol":   "$from_exchange",
			"price":    "$exchange_rate",
			"market":   "Forex",
			"currency": "$to_exchange",
		}}

		cursor, err = db.ExchangeCollection.Aggregate(context.TODO(), bson.A{project})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": tType,
			}).Error("failed to aggregate exchanges: %w", err)
			return nil, fmt.Errorf("failed to fetch exchanges")
		}
	} else {
		match := bson.M{"$match": bson.M{
			"_id.type":   tType,
			"_id.market": market,
		}}
		project := bson.M{"$project": bson.M{
			"name":     "$name",
			"symbol":   "$_id.symbol",
			"price":    "$price",
			"market":   "$_id.market",
			"currency": "$_id.stock_currency",
		}}

		cursor, err = db.InvestingCollection.Aggregate(context.TODO(), bson.A{match, project})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": tType,
			}).Error("failed to aggregate investings: %w", err)
			return nil, fmt.Errorf("failed to fetch investings")
		}
	}

	var investings []responses.InvestingTableResponse
	if err = cursor.All(context.TODO(), &investings); err != nil {
		logrus.WithFields(logrus.Fields{
			"type": tType,
		}).Error("failed to decode investings: %w", err)
		return nil, fmt.Errorf("failed to decode investings")
	}

	return investings, nil
}
