package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// var db.SubscriptionCollection = db.Database.Collection("subscriptions")
// var cardCollection = db.Database.Collection("cards")

type Subscription struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	CardID      *string            `bson:"card_id" json:"card_id"`
	Name        string             `bson:"name" json:"name"`
	Description *string            `bson:"description" json:"description"`
	BillDate    time.Time          `bson:"bill_date" json:"bill_date"`
	BillCycle   *BillCycle         `bson:"bill_cycle" json:"bill_cycle"`
	Price       float32            `bson:"price" json:"price"`
	Currency    string             `bson:"currency" json:"currency"`
	CreatedAt   time.Time          `bson:"created_at" json:"-"`
}

type Card struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Name       string             `bson:"name" json:"name"`
	Last4Digit string             `bson:"last_digit" json:"last_digit"`
	CardType   string             `bson:"type" json:"type"`
}

type BillCycle struct {
	Day   int `bson:"day" json:"day"`
	Month int `bson:"month" json:"month"`
	Year  int `bson:"year" json:"year"`
}

func createSubscriptionObject(uid, cardID, name, description, currency string, price float32, billDate time.Time, billCycle BillCycle) *Subscription {
	return &Subscription{
		UserID:      uid,
		CardID:      &cardID,
		Name:        name,
		Description: &description,
		BillDate:    billDate,
		BillCycle:   &billCycle,
		Price:       price,
		Currency:    currency,
		CreatedAt:   time.Now().UTC(),
	}
}

func createCardObject(uid, name, last4Digit, cardType string) *Card {
	return &Card{
		UserID:     uid,
		Name:       name,
		Last4Digit: last4Digit,
		CardType:   cardType,
	}
}

func CreateSubscription(data requests.Subscription) error {
	subscription := createSubscriptionObject(
		data.UserID,
		*data.CardID,
		data.Name,
		*data.Description,
		data.Currency,
		data.Price,
		data.BillDate,
		BillCycle(*data.BillCycle),
	)

	if _, err := db.SubscriptionCollection.InsertOne(context.TODO(), subscription); err != nil {
		return fmt.Errorf("failed to create new subscription: %w", err)
	}

	return nil
}

func CreateCard(data requests.Card) error {
	card := createCardObject(data.UserID, data.Name, data.Last4Digit, data.CardType)

	if _, err := db.CardCollection.InsertOne(context.TODO(), card); err != nil {
		return fmt.Errorf("failed to create new card: %w", err)
	}

	return nil
}

func GetCardsByUserID(uid string) ([]Card, error) {

	var cards []Card

	return cards, nil
}

//Sort by total payment if possible
func GetSubscriptionsByCardID(uid, cardID string) error {

	return nil
}

//Sort by name, closest bill date(?), price monthly
func GetSubscriptionsByUserID(uid string) ([]Subscription, error) {

	var subscriptions []Subscription

	return subscriptions, nil
}

func GetSubscriptionDetails(uid, subscriptionID string) (Subscription, error) {

	var subscription Subscription

	return subscription, nil
}

func UpdateSubscription(data requests.SubscriptionUpdate) error {
	return nil
}

// func UpdateCard(data requests.CardUpdate) error {
// 	return nil
// }

func DeleteSubscriptionBySubscriptionID(subscriptionID string) error {

	return nil
}

func DeleteSubscriptionsByUserID(uid string) error {

	return nil
}

func DeleteCardByCardID(cardID string) error {

	return nil
}

func DeleteCardsByUserID(uid string) error {

	return nil
}
