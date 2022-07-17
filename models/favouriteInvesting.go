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

		return fmt.Errorf("Failed to create new favourite investing.")
	}

	return nil
}

func (favInvestingModel *FavouriteInvestingModel) GetFavouriteInvestings(uid string) ([]responses.FavouriteInvesting, error) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}

	cursor, err := favInvestingModel.Collection.Aggregate(context.TODO(), bson.A{
		match,
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

		return nil, fmt.Errorf("Failed to decode favourite investings.")
	}

	return favInvestings, nil
}
