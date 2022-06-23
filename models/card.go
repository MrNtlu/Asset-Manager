package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CardModel struct {
	Collection *mongo.Collection
}

func NewCardModel(mongoDB *db.MongoDB) *CardModel {
	return &CardModel{
		Collection: mongoDB.Database.Collection("cards"),
	}
}

type Card struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Name       string             `bson:"name" json:"name"`
	Last4Digit string             `bson:"last_digit" json:"last_digit"`
	CardHolder string             `bson:"card_holder" json:"card_holder"`
	Color      string             `bson:"color" json:"color"`
	CardType   string             `bson:"type" json:"type"`
	Currency   string             `bson:"currency" json:"currency"`
	CreatedAt  time.Time          `bson:"created_at" json:"-"`
}

const cardPremiumLimit = 3

func createCardObject(uid, name, last4Digit, cardHolder, color, cardType, currency string) *Card {
	return &Card{
		UserID:     uid,
		Name:       name,
		Last4Digit: last4Digit,
		CardHolder: cardHolder,
		Color:      color,
		CardType:   cardType,
		Currency:   currency,
		CreatedAt:  time.Now().UTC(),
	}
}

func (cardModel *CardModel) CreateCard(uid string, data requests.Card) (Card, error) {
	card := createCardObject(uid, data.Name, data.Last4Digit, data.CardHolder, data.Color, data.CardType, data.Currency)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = cardModel.Collection.InsertOne(context.TODO(), card); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to create new card: ", err)

		return Card{}, fmt.Errorf("Failed to create new card.")
	}

	card.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *card, nil
}

func (cardModel *CardModel) GetUserCardCount(uid string) int64 {
	count, err := cardModel.Collection.CountDocuments(context.TODO(), bson.M{"user_id": uid})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to count user cards: ", err)

		return cardPremiumLimit
	}

	return count
}

func (cardModel *CardModel) GetCardByID(cardID string) (Card, error) {
	objectCardID, _ := primitive.ObjectIDFromHex(cardID)

	result := cardModel.Collection.FindOne(context.TODO(), bson.M{"_id": objectCardID})

	var card Card
	if err := result.Decode(&card); err != nil {
		logrus.WithFields(logrus.Fields{
			"card_id": cardID,
		}).Error("failed to find card by card id: ", err)

		return Card{}, fmt.Errorf("Failed to find card by card id.")
	}

	return card, nil
}

func (cardModel *CardModel) GetCardsByUserID(uid string) ([]Card, error) {
	match := bson.M{
		"user_id": uid,
	}
	sort := bson.M{
		"created_at": 1,
	}
	options := options.Find().SetSort(sort)

	cursor, err := cardModel.Collection.Find(context.TODO(), match, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find card by user id: ", err)

		return nil, fmt.Errorf("Failed to find card by user id.")
	}

	var cards []Card
	if err := cursor.All(context.TODO(), &cards); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode card: ", err)

		return nil, fmt.Errorf("Failed to decode card.")
	}

	return cards, nil
}

func (cardModel *CardModel) UpdateCard(data requests.CardUpdate, card Card) (Card, error) {
	objectCardID, _ := primitive.ObjectIDFromHex(data.ID)

	if data.Last4Digit != nil {
		card.Last4Digit = *data.Last4Digit
	}

	if data.Name != nil {
		card.Name = *data.Name
	}

	if data.CardHolder != nil {
		card.CardHolder = *data.CardHolder
	}

	if data.CardType != nil {
		card.CardType = *data.CardType
	}

	if data.Color != nil {
		card.Color = *data.Color
	}

	if data.Currency != nil {
		card.Currency = *data.Currency
	}

	if _, err := cardModel.Collection.UpdateOne(context.TODO(), bson.M{
		"_id": objectCardID,
	}, bson.M{"$set": card}); err != nil {
		logrus.WithFields(logrus.Fields{
			"card_id": data.ID,
			"data":    data,
		}).Error("failed to update card: ", err)

		return Card{}, fmt.Errorf("Failed to update card.")
	}

	return card, nil
}

func (cardModel *CardModel) DeleteCardByCardID(uid, cardID string) (bool, error) {
	objectCardID, _ := primitive.ObjectIDFromHex(cardID)

	count, err := cardModel.Collection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectCardID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":     uid,
			"card_id": cardID,
		}).Error("failed to delete card by card id: ", err)

		return false, fmt.Errorf("Failed to delete card by card id.")
	}

	return count.DeletedCount > 0, nil
}

func (cardModel *CardModel) DeleteAllCardsByUserID(uid string) error {
	if _, err := cardModel.Collection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all cards by user id: ", err)

		return fmt.Errorf("Failed to delete all cards by user id.")
	}

	return nil
}
