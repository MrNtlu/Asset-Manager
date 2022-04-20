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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Card struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Name       string             `bson:"name" json:"name"`
	Last4Digit string             `bson:"last_digit" json:"last_digit"`
	CardHolder string             `bson:"card_holder" json:"card_holder"`
	Color      string             `bson:"color" json:"color"`
	CardType   string             `bson:"type" json:"type"`
	CreatedAt  time.Time          `bson:"created_at" json:"-"`
}

func createCardObject(uid, name, last4Digit, cardHolder, color, cardType string) *Card {
	return &Card{
		UserID:     uid,
		Name:       name,
		Last4Digit: last4Digit,
		CardHolder: cardHolder,
		Color:      color,
		CardType:   cardType,
		CreatedAt:  time.Now().UTC(),
	}
}

func CreateCard(uid string, data requests.Card) error {
	card := createCardObject(uid, data.Name, data.Last4Digit, data.CardHolder, data.Color, data.CardType)

	if _, err := db.CardCollection.InsertOne(context.TODO(), card); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to create new card: ", err)
		return fmt.Errorf("failed to create new card")
	}

	return nil
}

func GetCardByID(cardID string) (Card, error) {
	objectCardID, _ := primitive.ObjectIDFromHex(cardID)

	result := db.CardCollection.FindOne(context.TODO(), bson.M{"_id": objectCardID})

	var card Card
	if err := result.Decode(&card); err != nil {
		logrus.WithFields(logrus.Fields{
			"card_id": cardID,
		}).Error("failed to find card by card id: ", err)
		return Card{}, fmt.Errorf("failed to find card by card id")
	}

	return card, nil
}

func GetCardsByUserID(uid string) ([]Card, error) {
	match := bson.M{
		"user_id": uid,
	}
	sort := bson.M{
		"created_at": 1,
	}
	options := options.Find().SetSort(sort)

	cursor, err := db.CardCollection.Find(context.TODO(), match, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find card by user id: ", err)
		return nil, fmt.Errorf("failed to find card by user id")
	}

	var cards []Card
	if err := cursor.All(context.TODO(), &cards); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode card: ", err)
		return nil, fmt.Errorf("failed to decode card")
	}

	return cards, nil
}

func UpdateCard(data requests.CardUpdate, card Card) error {
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

	if _, err := db.CardCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectCardID,
	}, bson.M{"$set": card}); err != nil {
		logrus.WithFields(logrus.Fields{
			"card_id": data.ID,
			"data":    data,
		}).Error("failed to update card: ", err)
		return fmt.Errorf("failed to update card")
	}

	return nil
}

func DeleteCardByCardID(uid, cardID string) (bool, error) {
	objectCardID, _ := primitive.ObjectIDFromHex(cardID)

	count, err := db.CardCollection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectCardID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":     uid,
			"card_id": cardID,
		}).Error("failed to delete card by card id: ", err)
		return false, fmt.Errorf("failed to delete card by card id")
	}

	return count.DeletedCount > 0, nil
}

func DeleteAllCardsByUserID(uid string) error {
	if _, err := db.CardCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all cards by user id: ", err)
		return fmt.Errorf("failed to delete all cards by user id")
	}

	return nil
}
