package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FavouriteInvestingModel struct {
	Collection *mongo.Collection
}

func NewFavouriteInvestingModel(mongoDB *db.MongoDB) *FavouriteInvestingModel {
	return &FavouriteInvestingModel{
		Collection: mongoDB.Database.Collection("favourite_investings"),
	}
}

type FavouriteInvesting struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	UserID      string               `bson:"user_id" json:"user_id"`
	InvestingID FavouriteInvestingID `bson:"investing_id" json:"investing_id"`
	Priority    int                  `bson:"priority" json:"priority"`
}

type FavouriteInvestingID struct {
	Symbol string `bson:"symbol" json:"symbol"`
	Type   string `bson:"type" json:"type"`
	Market string `bson:"market" json:"market"`
}

func createFavouriteInvesting(uid string, investingID FavouriteInvestingID, priority int) *FavouriteInvesting {
	return &FavouriteInvesting{
		UserID:      uid,
		InvestingID: investingID,
		Priority:    priority,
	}
}

func createFavouriteInvestingID(symbol, tType, market string) *FavouriteInvestingID {
	return &FavouriteInvestingID{
		Symbol: symbol,
		Type:   tType,
		Market: market,
	}
}

func (favInvestingModel *FavouriteInvestingModel) CreateFavouriteInvesting(uid string, data requests.FavouriteInvestingCreate) error {
	favInvesting := createFavouriteInvesting(
		uid,
		*createFavouriteInvestingID(data.Symbol, data.Type, data.Market),
		data.Priority,
	)

	if _, err := favInvestingModel.Collection.InsertOne(context.TODO(), favInvesting); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to create new favourite investing: ", err)

		return fmt.Errorf("Failed to create new watchlist.")
	}

	return nil
}

func (favInvestingModel *FavouriteInvestingModel) GetFavouriteInvestingsCount(uid string) int64 {
	count, err := favInvestingModel.Collection.CountDocuments(context.TODO(), bson.M{"user_id": uid})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to count user favourite investings: ", err)

		return subscriptionPremiumLimit
	}

	return count
}

func (favInvestingModel *FavouriteInvestingModel) GetFavouriteInvestings(uid string) ([]responses.FavouriteInvesting, error) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}
	lookup := bson.M{"$lookup": bson.M{
		"from": "investings",
		"let": bson.M{
			"symbol": "$investing_id.symbol",
			"type":   "$investing_id.type",
			"market": "$investing_id.market",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$_id.symbol", "$$symbol"}},
							bson.M{"$eq": bson.A{"$_id.type", "$$type"}},
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
		"preserveNullAndEmptyArrays": false,
	}}
	addCurrencyField := bson.M{"$addFields": bson.M{
		"price": "$investing.price",
		"currency": bson.M{
			"$ifNull": bson.A{
				"$investing._id.stock_currency",
				"USD",
			},
		},
	}}
	sort := bson.M{"$sort": bson.M{
		"priority": 1,
	}}

	cursor, err := favInvestingModel.Collection.Aggregate(context.TODO(), bson.A{
		match, lookup, unwindInvesting, addCurrencyField, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate favourite investings: ", err)

		return nil, fmt.Errorf("Failed to aggregate favourite investings.")
	}

	var favInvestings []responses.FavouriteInvesting
	if err = cursor.All(context.TODO(), &favInvestings); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode favourite investings: ", err)

		return nil, fmt.Errorf("Failed to decode watchlist.")
	}

	return favInvestings, nil
}

func (favInvestingModel *FavouriteInvestingModel) DeleteFavouriteInvestingByID(uid, fiID string) (bool, error) {
	objectFavInvestingID, _ := primitive.ObjectIDFromHex(fiID)

	count, err := favInvestingModel.Collection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectFavInvestingID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":                    uid,
			"favourite_investing_id": fiID,
		}).Error("failed to delete favourite investing by fav investing id: ", err)

		return false, fmt.Errorf("Failed to delete watchlist by id.")
	}

	return count.DeletedCount > 0, nil
}

func (favInvestingModel *FavouriteInvestingModel) DeleteAllFavouriteInvestingsByUserID(uid string) error {
	if _, err := favInvestingModel.Collection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all favourite investings by user id: ", err)

		return fmt.Errorf("Failed to delete all watchlist by user id.")
	}

	return nil
}
