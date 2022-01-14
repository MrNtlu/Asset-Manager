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

func GetSubscriptionByID(subscriptionID string) (Subscription, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	result := db.SubscriptionCollection.FindOne(context.TODO(), bson.M{"_id": objectSubscriptionID})

	var subscription Subscription
	if err := result.Decode(&subscription); err != nil {
		return Subscription{}, fmt.Errorf("failed to find subscription by subscription id: %w", err)
	}

	return subscription, nil
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
			"$let": bson.M{
				"vars": bson.M{
					"date_diff": bson.M{
						"$dateDiff": bson.M{
							"startDate": "$bill_date",
							"endDate":   time.Now(),
							"unit":      "day",
						},
					},
				},
				"in": bson.M{
					"$round": bson.A{
						bson.M{
							"$multiply": bson.A{
								bson.M{
									"$sum": bson.A{
										bson.M{
											"$floor": bson.M{
												"$divide": bson.A{
													bson.M{
														"$cond": bson.A{
															bson.M{
																"$gte": bson.A{
																	"$$date_diff",
																	1,
																},
															},
															"$$date_diff",
															-1,
														},
													}, "$bill_cycle",
												},
											},
										},
										1,
									},
								},
								"$price",
							},
						}, 1,
					},
				},
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

func GetSubscriptionStatisticsByUserID(uid string) ([]responses.SubscriptionStatistics, error) {
	match := bson.M{"$match": bson.M{
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
			"$let": bson.M{
				"vars": bson.M{
					"date_diff": bson.M{
						"$dateDiff": bson.M{
							"startDate": "$bill_date",
							"endDate":   time.Now(),
							"unit":      "day",
						},
					},
				},
				"in": bson.M{
					"$round": bson.A{
						bson.M{
							"$multiply": bson.A{
								bson.M{
									"$sum": bson.A{
										bson.M{
											"$floor": bson.M{
												"$divide": bson.A{
													bson.M{
														"$cond": bson.A{
															bson.M{
																"$gte": bson.A{
																	"$$date_diff",
																	1,
																},
															},
															"$$date_diff",
															-1,
														},
													}, "$bill_cycle",
												},
											},
										},
										1,
									},
								},
								"$price",
							},
						}, 1,
					},
				},
			},
		},
	}}
	sort := bson.M{"$sort": bson.M{
		"monthly_payment": -1,
	}}
	group := bson.M{"$group": bson.M{
		"_id": "$currency",
		"total_monthly_payment": bson.M{
			"$sum": "$monthly_payment",
		},
		"total_payment": bson.M{
			"$sum": "$total_payment",
		},
		"most_expensive": bson.M{
			"$first": "$monthly_payment",
		},
		"most_expensive_name": bson.M{
			"$first": "$name",
		},
	}}

	cursor, err := db.SubscriptionCollection.Aggregate(context.TODO(), bson.A{
		match, addFields, sort, group,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate subscription: %w", err)
	}

	var subscriptionStats []responses.SubscriptionStatistics
	if err = cursor.All(context.TODO(), &subscriptionStats); err != nil {
		return nil, fmt.Errorf("failed to decode subscriptions: %w", err)
	}

	return subscriptionStats, nil
}

func GetCardStatisticsByUserID(uid string) ([]responses.CardStatistics, error) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}
	set := bson.M{"$set": bson.M{
		"card_id": bson.M{
			"$toObjectId": "$card_id",
		},
	}}
	lookup := bson.M{"$lookup": bson.M{
		"from":         "cards",
		"localField":   "card_id",
		"foreignField": "_id",
		"as":           "card",
	}}
	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$card",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
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
			"$let": bson.M{
				"vars": bson.M{
					"date_diff": bson.M{
						"$dateDiff": bson.M{
							"startDate": "$bill_date",
							"endDate":   time.Now(),
							"unit":      "day",
						},
					},
				},
				"in": bson.M{
					"$round": bson.A{
						bson.M{
							"$multiply": bson.A{
								bson.M{
									"$sum": bson.A{
										bson.M{
											"$floor": bson.M{
												"$divide": bson.A{
													bson.M{
														"$cond": bson.A{
															bson.M{
																"$gte": bson.A{
																	"$$date_diff",
																	1,
																},
															},
															"$$date_diff",
															-1,
														},
													}, "$bill_cycle",
												},
											},
										},
										1,
									},
								},
								"$price",
							},
						}, 1,
					},
				},
			},
		},
	}}
	sort := bson.M{"$sort": bson.M{
		"monthly_payment": -1,
	}}
	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"card_id":  "$card_id",
			"currency": "$currency",
		},
		"currency": bson.M{
			"$first": "$currency",
		},
		"card_name": bson.M{
			"$first": "$card.name",
		},
		"card_last_digit": bson.M{
			"$first": "$card.last_digit",
		},
		"total_monthly_payment": bson.M{
			"$sum": "$monthly_payment",
		},
		"total_payment": bson.M{
			"$sum": "$total_payment",
		},
		"most_expensive": bson.M{
			"$first": "$monthly_payment",
		},
		"most_expensive_name": bson.M{
			"$first": "$name",
		},
	}}

	cursor, err := db.SubscriptionCollection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwind, addFields, sort, group,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate subscription: %w", err)
	}

	var cardStats []responses.CardStatistics
	if err = cursor.All(context.TODO(), &cardStats); err != nil {
		return nil, fmt.Errorf("failed to decode subscriptions: %w", err)
	}

	return cardStats, nil
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
	subscription.CardID = data.CardID

	if _, err := db.SubscriptionCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectSubscriptionID,
	}, bson.M{"$set": subscription}); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func UpdateSubscriptionCardIDToNull(uid string, cardID *string) {
	var match bson.M
	if cardID != nil {
		match = bson.M{
			"card_id": cardID,
			"user_id": uid,
		}
	} else {
		match = bson.M{
			"user_id": uid,
		}
	}

	if _, err := db.SubscriptionCollection.UpdateMany(context.TODO(), match,
		bson.M{"$set": bson.M{
			"card_id": nil,
		}}); err != nil {
		return
	}
}

func DeleteSubscriptionBySubscriptionID(uid, subscriptionID string) (bool, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	count, err := db.SubscriptionCollection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectSubscriptionID,
		"user_id": uid,
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete subscription: %w", err)
	}

	return count.DeletedCount > 0, nil
}

func DeleteAllSubscriptionsByUserID(uid string) error {
	if _, err := db.SubscriptionCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		return fmt.Errorf("failed to delete all subscriptions by user id: %w", err)
	}

	return nil
}
