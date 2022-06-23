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

type InvestingModel struct {
	Collection         *mongo.Collection
	ExchangeCollection *mongo.Collection
}

func NewInvestingModel(mongoDB *db.MongoDB) *InvestingModel {
	return &InvestingModel{
		Collection:         mongoDB.Database.Collection("investings"),
		ExchangeCollection: mongoDB.Database.Collection("exchanges"),
	}
}

func (investingModel *InvestingModel) GetInvestingsByTypeAndMarket(tType, market string) ([]responses.InvestingResponse, error) {
	match := bson.M{"$match": bson.M{
		"_id.type":   tType,
		"_id.market": market,
	}}
	project := bson.M{"$project": bson.M{
		"name":   "$name",
		"symbol": "$_id.symbol",
	}}

	cursor, err := investingModel.Collection.Aggregate(context.TODO(), bson.A{match, project})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": tType,
		}).Error("failed to aggregate investings: %w", err)

		return nil, fmt.Errorf("Failed to fetch investings.")
	}

	var investings []responses.InvestingResponse
	if err = cursor.All(context.TODO(), &investings); err != nil {
		logrus.WithFields(logrus.Fields{
			"type": tType,
		}).Error("failed to decode investings: %w", err)

		return nil, fmt.Errorf("Failed to decode investings.")
	}

	return investings, nil
}

func (investingModel *InvestingModel) GetInvestingPriceTableByTypeAndMarket(tType, market string) ([]responses.InvestingTableResponse, error) {
	var (
		cursor *mongo.Cursor
		err    error
	)

	if tType == "exchange" {
		project := bson.M{"$project": bson.M{
			"name":     "$from_exchange",
			"symbol":   "$from_exchange",
			"price":    "$exchange_rate",
			"market":   "Forex",
			"currency": "$to_exchange",
		}}

		cursor, err = investingModel.ExchangeCollection.Aggregate(context.TODO(), bson.A{project})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": tType,
			}).Error("failed to aggregate exchanges: %w", err)

			return nil, fmt.Errorf("Failed to fetch exchanges.")
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

		cursor, err = investingModel.Collection.Aggregate(context.TODO(), bson.A{match, project})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": tType,
			}).Error("failed to aggregate investings: %w", err)

			return nil, fmt.Errorf("Failed to fetch investings.")
		}
	}

	var investings []responses.InvestingTableResponse
	if err = cursor.All(context.TODO(), &investings); err != nil {
		logrus.WithFields(logrus.Fields{
			"type": tType,
		}).Error("failed to decode investings: %w", err)

		return nil, fmt.Errorf("Failed to decode investings.")
	}

	return investings, nil
}
