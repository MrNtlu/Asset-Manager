package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Card struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Name       string             `bson:"name" json:"name"`
	Last4Digit string             `bson:"last_digit" json:"last_digit"`
	CreatedAt  time.Time          `bson:"created_at" json:"-"`
}

func createCardObject(uid, name, last4Digit string) *Card {
	return &Card{
		UserID:     uid,
		Name:       name,
		Last4Digit: last4Digit,
		CreatedAt:  time.Now().UTC(),
	}
}

func CreateCard(uid string, data requests.Card) error {
	card := createCardObject(uid, data.Name, data.Last4Digit)

	if _, err := db.CardCollection.InsertOne(context.TODO(), card); err != nil {
		return fmt.Errorf("failed to create new card: %w", err)
	}

	return nil
}

func GetCardByID(cardID string) (Card, error) {
	objectCardID, _ := primitive.ObjectIDFromHex(cardID)

	result := db.CardCollection.FindOne(context.TODO(), bson.M{"_id": objectCardID})

	var card Card
	if err := result.Decode(&card); err != nil {
		return Card{}, fmt.Errorf("failed to find card by card id: %w", err)
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
		return nil, fmt.Errorf("failed to find card: %w", err)
	}

	var cards []Card
	if err := cursor.All(context.TODO(), &cards); err != nil {
		return nil, fmt.Errorf("failed to decode card: %w", err)
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

	if _, err := db.CardCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectCardID,
	}, bson.M{"$set": card}); err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	return nil
}

//TODO: Update card id of subscriptinos to null
func DeleteCardByCardID(cardID string) error {
	objectCardID, _ := primitive.ObjectIDFromHex(cardID)

	if _, err := db.CardCollection.DeleteOne(context.TODO(), bson.M{"_id": objectCardID}); err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}

	return nil
}

//TODO: Update all card id's of subscriptions to null
func DeleteAllCardsByUserID(uid string) error {
	if _, err := db.CardCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		return fmt.Errorf("failed to delete all cards by user id: %w", err)
	}

	return nil
}
