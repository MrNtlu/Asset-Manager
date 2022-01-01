package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	BillCycle   *int               `bson:"bill_cycle" json:"bill_cycle"`
	Price       float64            `bson:"price" json:"price"`
	Currency    string             `bson:"currency" json:"currency"`
	CreatedAt   time.Time          `bson:"created_at" json:"-"`
}

type Card struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Name       string             `bson:"name" json:"name"`
	Last4Digit string             `bson:"last_digit" json:"last_digit"`
	CreatedAt  time.Time          `bson:"created_at" json:"-"`
}

func createSubscriptionObject(uid, name, currency string, cardID, description *string, price float64, billDate time.Time, billCycle *int) *Subscription {
	return &Subscription{
		UserID:      uid,
		CardID:      cardID,
		Name:        name,
		Description: description,
		BillDate:    billDate,
		BillCycle:   billCycle,
		Price:       price,
		Currency:    currency,
		CreatedAt:   time.Now().UTC(),
	}
}

func createCardObject(uid, name, last4Digit string) *Card {
	return &Card{
		UserID:     uid,
		Name:       name,
		Last4Digit: last4Digit,
		CreatedAt:  time.Now().UTC(),
	}
}

func CreateSubscription(uid string, data requests.Subscription) error {
	subscription := createSubscriptionObject(
		uid,
		data.Name,
		data.Currency,
		data.CardID,
		data.Description,
		data.Price,
		data.BillDate,
		data.BillCycle,
	)

	if _, err := db.SubscriptionCollection.InsertOne(context.TODO(), subscription); err != nil {
		return fmt.Errorf("failed to create new subscription: %w", err)
	}

	return nil
}

func CreateCard(uid string, data requests.Card) error {
	card := createCardObject(uid, data.Name, data.Last4Digit)

	if _, err := db.CardCollection.InsertOne(context.TODO(), card); err != nil {
		return fmt.Errorf("failed to create new card: %w", err)
	}

	return nil
}

func GetSubscriptionByID(subscriptionID string) (Subscription, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	result := db.SubscriptionCollection.FindOne(context.TODO(), bson.M{"_id": objectSubscriptionID})

	var subscription Subscription
	if err := result.Decode(&subscription); err != nil {
		return Subscription{}, fmt.Errorf("failed to find subscription by subscription id: %w", err)
	}

	return subscription, nil
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

func GetSubscriptionsByCardID(uid, cardID string) ([]Subscription, error) {
	match := bson.M{
		"card_id": cardID,
		"user_id": uid,
	}
	sort := bson.M{
		"name": 1,
	}
	options := options.Find().SetSort(sort)

	cursor, err := db.SubscriptionCollection.Find(context.TODO(), match, options)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	var subscriptions []Subscription
	if err := cursor.All(context.TODO(), &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to decode subscription: %w", err)
	}

	return subscriptions, nil
}

func GetSubscriptionsByUserID(uid string, data requests.SubscriptionSort) ([]Subscription, error) {
	match := bson.M{
		"user_id": uid,
	}

	var sort bson.D
	if data.Sort == "price" {
		sort = bson.D{
			primitive.E{Key: "currency", Value: 1},
			primitive.E{Key: data.Sort, Value: data.SortType},
		}
	} else {
		sort = bson.D{
			{Key: data.Sort, Value: data.SortType},
		}
	}
	options := options.Find().SetSort(sort)

	cursor, err := db.SubscriptionCollection.Find(context.TODO(), match, options)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	var subscriptions []Subscription
	if err := cursor.All(context.TODO(), &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to decode subscription: %w", err)
	}

	return subscriptions, nil
}

func GetSubscriptionDetails(uid, subscriptionID string) (responses.SubscriptionDetails, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	match := bson.M{"$match": bson.M{
		"_id":     objectSubscriptionID,
		"user_id": uid,
	}}
	addFields := bson.M{"$addFields": bson.M{
		"monthly_payment": bson.M{
			"$round": bson.A{
				bson.M{
					"$multiply": bson.A{
						bson.M{"$divide": bson.A{30, "$bill_cycle"}},
						"$price",
					},
				}, 1,
			},
		},
		"total_payment": bson.M{
			"$round": bson.A{
				bson.M{
					"$multiply": bson.A{
						bson.M{
							"$floor": bson.M{
								"$divide": bson.A{
									bson.M{
										"$dateDiff": bson.M{
											"startDate": "$bill_date",
											"endDate":   time.Now(),
											"unit":      "day",
										},
									}, "$bill_cycle",
								},
							},
						},
						"$price",
					},
				}, 1,
			},
		},
	}}

	cursor, err := db.SubscriptionCollection.Aggregate(context.TODO(), bson.A{match, addFields})
	if err != nil {
		return responses.SubscriptionDetails{}, fmt.Errorf("failed to aggregate subscription: %w", err)
	}

	var subscriptions []responses.SubscriptionDetails
	if err = cursor.All(context.TODO(), &subscriptions); err != nil {
		return responses.SubscriptionDetails{}, fmt.Errorf("failed to decode subscriptions: %w", err)
	}

	if len(subscriptions) > 0 {
		return subscriptions[0], nil
	}

	return responses.SubscriptionDetails{}, nil
}

//TODO: Aggregation
func GetSubscriptionStatisticsByUserID(uid string) error {
	return nil
}

func UpdateSubscription(data requests.SubscriptionUpdate, subscription Subscription) error {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(data.ID)

	if data.Name != nil {
		subscription.Name = *data.Name
	}
	if data.Description != nil {
		subscription.Description = data.Description
	}
	if data.BillDate != nil {
		subscription.BillDate = *data.BillDate
	}
	if data.BillCycle != nil {
		subscription.BillCycle = data.BillCycle
	}
	if data.Price != nil {
		subscription.Price = *data.Price
	}
	if data.Currency != nil {
		subscription.Currency = *data.Currency
	}

	if _, err := db.CardCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectSubscriptionID,
	}, bson.M{"$set": subscription}); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
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

func DeleteSubscriptionBySubscriptionID(subscriptionID string) error {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	if _, err := db.CardCollection.DeleteOne(context.TODO(), bson.M{"_id": objectSubscriptionID}); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

func DeleteAllSubscriptionsByUserID(uid string) error {
	if _, err := db.SubscriptionCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		return fmt.Errorf("failed to delete all subscriptions by user id: %w", err)
	}

	return nil
}

func DeleteCardByCardID(cardID string) error {
	objectCardID, _ := primitive.ObjectIDFromHex(cardID)

	if _, err := db.CardCollection.DeleteOne(context.TODO(), bson.M{"_id": objectCardID}); err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}

	return nil
}

func DeleteAllCardsByUserID(uid string) error {
	if _, err := db.CardCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		return fmt.Errorf("failed to delete all cards by user id: %w", err)
	}

	return nil
}
