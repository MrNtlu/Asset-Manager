package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/teambition/rrule-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Subscription struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	CardID      *string            `bson:"card_id" json:"card_id"`
	Name        string             `bson:"name" json:"name"`
	Description *string            `bson:"description" json:"description"`
	BillDate    time.Time          `bson:"bill_date" json:"bill_date"`
	BillCycle   BillCycle          `bson:"bill_cycle" json:"bill_cycle"`
	Price       float64            `bson:"price" json:"price"`
	Currency    string             `bson:"currency" json:"currency"`
	Color       string             `bson:"color" json:"color"`
	Image       string             `bson:"image" json:"image"`
	CreatedAt   time.Time          `bson:"created_at" json:"-"`
}

type BillCycle struct {
	Day   int `bson:"day" json:"day"`
	Month int `bson:"month" json:"month"`
	Year  int `bson:"year" json:"year"`
}

func createSubscriptionObject(uid, name, currency, color, image string, cardID, description *string, price float64, billDate time.Time, billCycle BillCycle) *Subscription {
	return &Subscription{
		UserID:      uid,
		CardID:      cardID,
		Name:        name,
		Description: description,
		BillDate:    billDate,
		BillCycle:   billCycle,
		Price:       price,
		Currency:    currency,
		Color:       color,
		Image:       image,
		CreatedAt:   time.Now().UTC(),
	}
}

func createBillCycle(billCycle requests.BillCycle) *BillCycle {
	return &BillCycle{
		Day:   billCycle.Day,
		Month: billCycle.Month,
		Year:  billCycle.Year,
	}
}

func CreateSubscription(uid string, data requests.Subscription) (responses.Subscription, error) {
	subscription := createSubscriptionObject(
		uid,
		data.Name,
		data.Currency,
		data.Color,
		data.Image,
		data.CardID,
		data.Description,
		data.Price,
		data.BillDate,
		*createBillCycle(data.BillCycle),
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)
	if insertedID, err = db.SubscriptionCollection.InsertOne(context.TODO(), subscription); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new subscription: ", err)
		return responses.Subscription{}, fmt.Errorf("Failed to create new subscription.")
	}
	subscription.ID = insertedID.InsertedID.(primitive.ObjectID)

	return convertModelToResponse(*subscription), nil
}

func GetUserSubscriptionCount(uid string) int64 {
	count, err := db.SubscriptionCollection.CountDocuments(context.TODO(), bson.M{"user_id": uid})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to count user subscriptions: ", err)
		return 5
	}

	return count
}

func GetSubscriptionByID(subscriptionID string) (Subscription, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	result := db.SubscriptionCollection.FindOne(context.TODO(), bson.M{"_id": objectSubscriptionID})

	var subscription Subscription
	if err := result.Decode(&subscription); err != nil {
		logrus.WithFields(logrus.Fields{
			"subscription_id": subscriptionID,
		}).Error("failed to create new subscription: ", err)
		return Subscription{}, fmt.Errorf("Failed to find subscription by subscription id.")
	}

	return subscription, nil
}

func GetSubscriptionsByCardID(uid, cardID string) ([]responses.Subscription, error) {
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
		logrus.WithFields(logrus.Fields{
			"uid":     uid,
			"card_id": cardID,
		}).Error("failed to find subscription: ", err)
		return nil, fmt.Errorf("Failed to find subscription.")
	}

	var subscriptions []responses.Subscription
	if err := cursor.All(context.TODO(), &subscriptions); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":     uid,
			"card_id": cardID,
		}).Error("failed to decode subscriptions: ", err)
		return nil, fmt.Errorf("Failed to decode subscriptions.")
	}

	for index, subscription := range subscriptions {
		subscriptions[index].NextBillDate = getNextBillDate(
			subscription.BillCycle,
			subscription.BillDate,
		)
	}

	return subscriptions, nil
}

func GetSubscriptionsByUserID(uid string, data requests.SubscriptionSort) ([]responses.Subscription, error) {
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
		logrus.WithFields(logrus.Fields{
			"uid":       uid,
			"sort":      data.Sort,
			"sort_type": data.SortType,
		}).Error("failed to find subscription: ", err)
		return nil, fmt.Errorf("Failed to find subscription.")
	}

	var subscriptions []responses.Subscription
	if err := cursor.All(context.TODO(), &subscriptions); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":       uid,
			"sort":      data.Sort,
			"sort_type": data.SortType,
		}).Error("failed to decode subscription: ", err)
		return nil, fmt.Errorf("Failed to decode subscription.")
	}

	for index, subscription := range subscriptions {
		subscriptions[index].NextBillDate = getNextBillDate(
			subscription.BillCycle,
			subscription.BillDate,
		)
	}

	return subscriptions, nil
}

func GetSubscriptionDetails(uid, subscriptionID string) (responses.SubscriptionDetails, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	match := bson.M{"$match": bson.M{
		"_id":     objectSubscriptionID,
		"user_id": uid,
	}}

	cursor, err := db.SubscriptionCollection.Aggregate(context.TODO(), bson.A{match, addSubscriptionMonthlyAndTotalPaymentFields()})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":             uid,
			"subscription_id": subscriptionID,
		}).Error("failed to aggregate subscription details: ", err)
		return responses.SubscriptionDetails{}, fmt.Errorf("Failed to aggregate subscription details.")
	}

	var subscriptions []responses.SubscriptionDetails
	if err = cursor.All(context.TODO(), &subscriptions); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":             uid,
			"subscription_id": subscriptionID,
		}).Error("failed to decode subscription details: ", err)
		return responses.SubscriptionDetails{}, fmt.Errorf("Failed to decode subscription details.")
	}

	if len(subscriptions) > 0 {
		subscription := subscriptions[0]
		subscription.NextBillDate = getNextBillDate(
			subscription.BillCycle,
			subscription.BillDate,
		)
		return subscription, nil
	}

	return responses.SubscriptionDetails{}, nil
}

func GetSubscriptionStatisticsByUserID(uid string) ([]responses.SubscriptionStatistics, error) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}
	group := bson.M{"$group": bson.M{
		"_id": "$currency",
		"total_monthly_payment": bson.M{
			"$sum": "$monthly_payment",
		},
		"total_payment": bson.M{
			"$sum": "$total_payment",
		},
	}}

	cursor, err := db.SubscriptionCollection.Aggregate(context.TODO(), bson.A{
		match, addSubscriptionMonthlyAndTotalPaymentFields(), group,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate subscription statistics: ", err)
		return nil, fmt.Errorf("Failed to aggregate subscription statistics.")
	}

	var subscriptionStats []responses.SubscriptionStatistics
	if err = cursor.All(context.TODO(), &subscriptionStats); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode subscription statistics: ", err)
		return nil, fmt.Errorf("Failed to decode subscription statistics.")
	}

	return subscriptionStats, nil
}

func GetCardStatisticsByUserIDAndCardID(uid, cardID string) (responses.CardSubscriptionStatistics, error) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
		"card_id": cardID,
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
	exchangeLookup := bson.M{"$lookup": bson.M{
		"from": "exchanges",
		"let": bson.M{
			"card_currency": "$card.currency",
			"sub_currency":  "$currency",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$cond": bson.A{
							bson.M{"$ne": bson.A{"$$card_currency", "$$sub_currency"}},
							bson.M{
								"$and": bson.A{
									bson.M{"$eq": bson.A{"$to_exchange", "$$card_currency"}},
									bson.M{"$eq": bson.A{"$from_exchange", "$$sub_currency"}},
								},
							},
							nil,
						},
					},
				},
			},
		},
		"as": "card_exchange_rate",
	}}
	unwindExchange := bson.M{"$unwind": bson.M{
		"path":                       "$card_exchange_rate",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	project := bson.M{"$project": bson.M{
		"bill_date":  true,
		"bill_cycle": true,
		"currency":   "$card.currency",
		"price": bson.M{
			"$ifNull": bson.A{
				bson.M{
					"$multiply": bson.A{
						"$price", "$card_exchange_rate.exchange_rate",
					},
				},
				"$price",
			},
		},
	}}
	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"card_id":  "$card_id",
			"currency": "$currency",
		},
		"currency": bson.M{
			"$first": "$currency",
		},
		"total_monthly_payment": bson.M{
			"$sum": "$monthly_payment",
		},
		"total_payment": bson.M{
			"$sum": "$total_payment",
		},
	}}

	cursor, err := db.SubscriptionCollection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwind, exchangeLookup, unwindExchange, project, addSubscriptionMonthlyAndTotalPaymentFields(), group,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate card statistics: ", err)
		return responses.CardSubscriptionStatistics{}, fmt.Errorf("Failed to aggregate card statistics: %w", err)
	}

	var cardStats []responses.CardSubscriptionStatistics
	if err = cursor.All(context.TODO(), &cardStats); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode card statistics: ", err)
		return responses.CardSubscriptionStatistics{}, fmt.Errorf("Failed to decode card statistics: %w", err)
	}

	if len(cardStats) > 0 {
		return cardStats[0], nil
	}

	return responses.CardSubscriptionStatistics{}, nil
}

func UpdateSubscription(data requests.SubscriptionUpdate, subscription Subscription) (responses.Subscription, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(data.ID)

	if data.Name != nil {
		subscription.Name = *data.Name
	}
	if data.Description != nil {
		subscription.Description = data.Description
	}
	if data.Color != nil {
		subscription.Color = *data.Color
	}
	if data.Image != nil {
		subscription.Image = *data.Image
	}
	if data.BillDate != nil {
		subscription.BillDate = *data.BillDate
	}
	if data.BillCycle != nil {
		subscription.BillCycle = *createBillCycle(*data.BillCycle)
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
		logrus.WithFields(logrus.Fields{
			"subscription_id": data.ID,
			"data":            data,
		}).Error("failed to update subscription: ", err)
		return responses.Subscription{}, fmt.Errorf("Failed to update subscription.")
	}

	return convertModelToResponse(subscription), nil
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
		logrus.WithFields(logrus.Fields{
			"uid":             uid,
			"subscription_id": subscriptionID,
		}).Error("failed to delete subscription: ", err)
		return false, fmt.Errorf("Failed to delete subscription.")
	}

	return count.DeletedCount > 0, nil
}

func DeleteAllSubscriptionsByUserID(uid string) error {
	if _, err := db.SubscriptionCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all subscriptions by user id: ", err)
		return fmt.Errorf("Failed to delete all subscriptions by user id.")
	}

	return nil
}

func getNextBillDate(billCycle responses.BillCycle, initialBillDate time.Time) time.Time {
	var (
		todayDate      = time.Now().UTC()
		freq           rrule.Frequency
		count          int
		comparisonDate time.Time
		billDate       time.Time
	)

	if billCycle.Day != 0 {
		freq = rrule.DAILY
		count = billCycle.Day
	} else if billCycle.Month != 0 {
		freq = rrule.MONTHLY
		count = billCycle.Month
	} else if billCycle.Year != 0 {
		freq = rrule.YEARLY
		count = billCycle.Year
	}

	comparisonDate = time.Date(todayDate.Year(), todayDate.Month(), todayDate.Day(), 0, 0, 0, 0, todayDate.Location())
	billDate = time.Date(initialBillDate.Year(), initialBillDate.Month(), initialBillDate.Day(), 23, 59, 59, 0, initialBillDate.Location())

	rule, _ := rrule.NewRRule(rrule.ROption{
		Freq:     freq,
		Interval: count,
		Dtstart:  billDate,
	})

	return rule.After(comparisonDate, true)
}

func convertModelToResponse(subscription Subscription) responses.Subscription {
	billCycle := responses.BillCycle{
		Day:   subscription.BillCycle.Day,
		Month: subscription.BillCycle.Month,
		Year:  subscription.BillCycle.Year,
	}
	return responses.Subscription{
		ID:          subscription.ID,
		UserID:      subscription.UserID,
		CardID:      subscription.CardID,
		Name:        subscription.Name,
		Description: subscription.Description,
		BillDate:    subscription.BillDate,
		NextBillDate: getNextBillDate(
			billCycle,
			subscription.BillDate,
		),
		BillCycle: billCycle,
		Price:     subscription.Price,
		Currency:  subscription.Currency,
		Color:     subscription.Color,
		Image:     &subscription.Image,
		CreatedAt: subscription.CreatedAt,
	}
}

func addSubscriptionMonthlyAndTotalPaymentFields() bson.M {
	return bson.M{"$addFields": bson.M{
		"monthly_payment": bson.M{
			"$round": bson.A{
				bson.M{
					"$switch": bson.M{
						"branches": bson.A{
							//Day case
							bson.M{
								"case": bson.M{"$gt": bson.A{"$bill_cycle.day", 0}},
								"then": bson.M{
									"$multiply": bson.A{
										bson.M{
											"$divide": bson.A{30, "$bill_cycle.day"},
										},
										"$price",
									},
								},
							},
							//Month Case
							bson.M{
								"case": bson.M{"$gt": bson.A{"$bill_cycle.month", 1}},
								"then": bson.M{
									"$divide": bson.A{"$price", "$bill_cycle.month"},
								},
							},
							//Year Case
							bson.M{
								"case": bson.M{"$gt": bson.A{"$bill_cycle.year", 0}},
								"then": bson.M{
									"$divide": bson.A{
										bson.M{
											"$multiply": bson.A{
												12,
												"$bill_cycle.year",
											},
										},
										"$price",
									},
								},
							},
						},
						"default": "$price",
					},
				},
				2,
			},
		},
		"total_payment": bson.M{
			"$let": bson.M{
				"vars": bson.M{
					"date_diff": bson.M{
						"$round": bson.M{
							"$divide": bson.A{
								bson.M{
									"$subtract": bson.A{time.Now(), "$bill_date"},
								},
								86400000,
							},
						},
					},
				},
				"in": bson.M{
					"$round": bson.A{
						bson.M{
							"$cond": bson.A{
								bson.M{
									"$gte": bson.A{"$$date_diff", 1},
								},
								bson.M{
									"$switch": bson.M{
										"branches": bson.A{
											//Day case
											bson.M{
												"case": bson.M{"$gt": bson.A{"$bill_cycle.day", 0}},
												"then": bson.M{
													"$multiply": bson.A{
														bson.M{
															"$ceil": bson.M{
																"$divide": bson.A{"$$date_diff", "$bill_cycle.day"},
															},
														},
														"$price",
													},
												},
											},
											//Month Case
											bson.M{
												"case": bson.M{"$gt": bson.A{"$bill_cycle.month", 0}},
												"then": bson.M{
													"$multiply": bson.A{
														bson.M{
															"$ceil": bson.M{
																"$divide": bson.A{
																	bson.M{
																		"$ceil": bson.M{
																			"$divide": bson.A{"$$date_diff", 30},
																		},
																	},
																	"$bill_cycle.month",
																},
															},
														},
														"$price",
													},
												},
											},
											//Year Case
											bson.M{
												"case": bson.M{"$gt": bson.A{"$bill_cycle.year", 0}},
												"then": bson.M{
													"$multiply": bson.A{
														bson.M{
															"$ceil": bson.M{
																"$divide": bson.A{
																	bson.M{
																		"$ceil": bson.M{
																			"$divide": bson.A{"$$date_diff", 365},
																		},
																	},
																	"$bill_cycle.year",
																},
															},
														},
														"$price",
													},
												},
											},
										},
										"default": "$price",
									},
								},
								0,
							},
						},
						2,
					},
				},
			},
		},
	}}
}
