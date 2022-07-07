package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"asset_backend/responses"
	"asset_backend/utils"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/teambition/rrule-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubscriptionModel struct {
	Collection       *mongo.Collection
	InviteCollection *mongo.Collection
}

func NewSubscriptionModel(mongoDB *db.MongoDB) *SubscriptionModel {
	return &SubscriptionModel{
		Collection:       mongoDB.Database.Collection("subscriptions"),
		InviteCollection: mongoDB.Database.Collection("subscription-invites"),
	}
}

type Subscription struct {
	ID               primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	UserID           string               `bson:"user_id" json:"user_id"`
	CardID           *string              `bson:"card_id" json:"card_id"`
	Name             string               `bson:"name" json:"name"`
	Description      *string              `bson:"description" json:"description"`
	BillDate         time.Time            `bson:"bill_date" json:"bill_date"`
	BillCycle        BillCycle            `bson:"bill_cycle" json:"bill_cycle"`
	Price            float64              `bson:"price" json:"price"`
	Currency         string               `bson:"currency" json:"currency"`
	Color            string               `bson:"color" json:"color"`
	Image            string               `bson:"image" json:"image"`
	Account          *SubscriptionAccount `bson:"account" json:"account"`
	SharedUsers      []string             `bson:"shared_users" json:"shared_users"`
	InvitedUsers     []string             `bson:"invited_users" json:"invited_users"`
	NotificationTime *time.Time           `bson:"notification_time" json:"notification_time"`
	CreatedAt        time.Time            `bson:"created_at" json:"-"`
}

type SubscriptionAccount struct {
	EmailAddress string  `bson:"email_address" json:"email_address"`
	Password     *string `bson:"password" json:"password"`
}

type SubscriptionInvite struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID         string             `bson:"user_id" json:"user_id"`
	InvitedUserID  string             `bson:"invited_user_id" json:"invited_user_id"`
	SubscriptionID string             `bson:"subscription_id" json:"subscription_id"`
	CreatedAt      time.Time          `bson:"created_at" json:"-"`
}

type BillCycle struct {
	Day   int `bson:"day" json:"day"`
	Month int `bson:"month" json:"month"`
	Year  int `bson:"year" json:"year"`
}

const subscriptionPremiumLimit = 5

func createSubscriptionObject(
	uid, name, currency, color, image string,
	cardID, description *string, price float64,
	billDate time.Time, billCycle BillCycle, account *SubscriptionAccount,
	notification *time.Time,
) *Subscription {
	return &Subscription{
		UserID:           uid,
		CardID:           cardID,
		Name:             name,
		Description:      description,
		BillDate:         billDate,
		BillCycle:        billCycle,
		Price:            price,
		Currency:         currency,
		Color:            color,
		Image:            image,
		SharedUsers:      make([]string, 0),
		InvitedUsers:     make([]string, 0),
		Account:          account,
		NotificationTime: notification,
		CreatedAt:        time.Now().UTC(),
	}
}

func createSubscriptionAccount(account *requests.SubscriptionAccount) *SubscriptionAccount {
	if account.Password != nil {
		hashedPassword := utils.Encrypt(*account.Password)

		return &SubscriptionAccount{
			EmailAddress: account.EmailAddress,
			Password:     &hashedPassword,
		}
	}

	return &SubscriptionAccount{
		EmailAddress: account.EmailAddress,
	}
}

func createBillCycle(billCycle requests.BillCycle) *BillCycle {
	return &BillCycle{
		Day:   billCycle.Day,
		Month: billCycle.Month,
		Year:  billCycle.Year,
	}
}

func createSubscriptionInvite(uid, invitedUID, subscriptionID string) *SubscriptionInvite {
	return &SubscriptionInvite{
		UserID:         uid,
		InvitedUserID:  invitedUID,
		SubscriptionID: subscriptionID,
		CreatedAt:      time.Now().UTC(),
	}
}

func (subscriptionModel *SubscriptionModel) CreateSubscription(uid string, data requests.Subscription) (responses.Subscription, error) {
	var subscriptionAccount *SubscriptionAccount
	if data.Account != nil {
		subscriptionAccount = createSubscriptionAccount(data.Account)
	}

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
		subscriptionAccount,
		data.NotificationTime,
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = subscriptionModel.Collection.InsertOne(context.TODO(), subscription); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new subscription: ", err)

		return responses.Subscription{}, fmt.Errorf("Failed to create new subscription.")
	}

	subscription.ID = insertedID.InsertedID.(primitive.ObjectID)

	return convertModelToResponse(*subscription), nil
}

func (subscriptionModel *SubscriptionModel) InviteSubscriptionToUser(uid, invitedUID, subscriptionID string) error {
	subscriptionInvite := createSubscriptionInvite(uid, invitedUID, subscriptionID)

	if isInviteSent := subscriptionModel.isInviteSentToUser(uid, invitedUID, subscriptionID); isInviteSent {
		return fmt.Errorf("Invitation already sent.")
	}

	if _, err := subscriptionModel.InviteCollection.InsertOne(context.TODO(), subscriptionInvite); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":             uid,
			"invited_uid":     invitedUID,
			"subscription_id": subscriptionID,
		}).Error("failed to invite user: ", err)

		return fmt.Errorf("Failed to invite user.")
	}

	if err := subscriptionModel.UpdateSubscriptionInvite(invitedUID, subscriptionID, true, false); err != nil {
		return err
	}

	return nil
}

func (subscriptionModel *SubscriptionModel) GetUserSubscriptionCount(uid string) int64 {
	count, err := subscriptionModel.Collection.CountDocuments(context.TODO(), bson.M{"user_id": uid})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to count user subscriptions: ", err)

		return subscriptionPremiumLimit
	}

	return count
}

func (subscriptionModel *SubscriptionModel) GetSubscriptionByID(subscriptionID string) (Subscription, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	result := subscriptionModel.Collection.FindOne(context.TODO(), bson.M{"_id": objectSubscriptionID})

	var subscription Subscription
	if err := result.Decode(&subscription); err != nil {
		logrus.WithFields(logrus.Fields{
			"subscription_id": subscriptionID,
		}).Error("failed to create new subscription: ", err)

		return Subscription{}, fmt.Errorf("Failed to find subscription by subscription id.")
	}

	return subscription, nil
}

func (subscriptionModel *SubscriptionModel) GetSubscriptionsByCardID(uid, cardID string) ([]responses.Subscription, error) {
	match := bson.M{
		"card_id": cardID,
		"user_id": uid,
	}
	sort := bson.M{
		"name": 1,
	}
	options := options.Find().SetSort(sort)

	cursor, err := subscriptionModel.Collection.Find(context.TODO(), match, options)
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

func (subscriptionModel *SubscriptionModel) GetSubscriptionsByUserID(
	uid string, data requests.SubscriptionSort, isSharedSubscriptions bool,
) ([]responses.Subscription, error) {
	var match bson.M

	if isSharedSubscriptions {
		match = bson.M{"shared_users": bson.M{
			"$in": bson.A{uid},
		}}
	} else {
		match = bson.M{
			"user_id": uid,
		}
	}

	var sortOptions bson.D
	if data.Sort == "price" {
		sortOptions = bson.D{
			primitive.E{Key: "currency", Value: 1},
			primitive.E{Key: data.Sort, Value: data.SortType},
		}
	} else {
		sortOptions = bson.D{
			{Key: data.Sort, Value: data.SortType},
		}
	}

	options := options.Find().SetSort(sortOptions)

	cursor, err := subscriptionModel.Collection.Find(context.TODO(), match, options)
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

		if subscription.Account != nil && subscription.Account.Password != nil {
			decryptedPassword := utils.Decrypt(*subscription.Account.Password)
			subscriptions[index].Account.Password = &decryptedPassword
		}
	}

	if data.Sort == "date" {
		sort.Slice(subscriptions, func(i, j int) bool {
			if data.SortType == -1 {
				return subscriptions[i].NextBillDate.Before(subscriptions[j].NextBillDate)
			} else {
				return subscriptions[i].NextBillDate.After(subscriptions[j].NextBillDate)
			}
		})
	}

	return subscriptions, nil
}

func (subscriptionModel *SubscriptionModel) GetSubscriptionDetails(uid, subscriptionID string) (responses.SubscriptionDetails, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	match := bson.M{"$match": bson.M{
		"_id":     objectSubscriptionID,
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
		"preserveNullAndEmptyArrays": true,
	}}

	cursor, err := subscriptionModel.Collection.Aggregate(context.TODO(), bson.A{match, addSubscriptionMonthlyAndTotalPaymentFields(), set, lookup, unwind})
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

		if subscription.Account != nil && subscription.Account.Password != nil {
			decryptedPassword := utils.Decrypt(*subscription.Account.Password)
			subscription.Account.Password = &decryptedPassword
		}

		return subscription, nil
	}

	return responses.SubscriptionDetails{}, nil
}

func (subscriptionModel *SubscriptionModel) GetSubscriptionStatisticsByUserID(uid string) ([]responses.SubscriptionStatistics, error) {
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

	cursor, err := subscriptionModel.Collection.Aggregate(context.TODO(), bson.A{
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

func (subscriptionModel *SubscriptionModel) GetCardStatisticsByUserIDAndCardID(uid, cardID string) (responses.CardSubscriptionStatistics, error) {
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

	cursor, err := subscriptionModel.Collection.Aggregate(context.TODO(), bson.A{
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

func (subscriptionModel *SubscriptionModel) UpdateSubscriptionInvite(invitedUID, subscriptionID string, isPush, isInvitationAccepted bool) error {
	objectID, _ := primitive.ObjectIDFromHex(subscriptionID)

	var operation bson.M
	if isPush {
		operation = bson.M{"$push": bson.M{
			"invited_users": invitedUID,
		}}
	} else {
		operation = bson.M{"$pull": bson.M{
			"invited_users": invitedUID,
		}}
	}

	if _, err := subscriptionModel.Collection.UpdateOne(context.TODO(), bson.M{"_id": objectID}, operation); err != nil {
		return fmt.Errorf("Failed to set subscription.")
	}

	if isInvitationAccepted {
		if _, err := subscriptionModel.Collection.UpdateOne(context.TODO(), bson.M{"_id": objectID}, bson.M{
			"$push": bson.M{
				"shared_users": invitedUID,
			},
		}); err != nil {
			return fmt.Errorf("Failed to set subscription.")
		}
	}

	return nil
}

func (subscriptionModel *SubscriptionModel) UpdateSubscription(data requests.SubscriptionUpdate, subscription Subscription) (responses.Subscription, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(data.ID)

	if data.Name != nil {
		subscription.Name = *data.Name
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

	subscription.NotificationTime = data.NotificationTime

	subscription.CardID = data.CardID

	subscription.Description = data.Description

	if data.Account != nil {
		subscription.Account = createSubscriptionAccount(data.Account)
	} else {
		subscription.Account = nil
	}

	if _, err := subscriptionModel.Collection.UpdateOne(context.TODO(), bson.M{
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

func (subscriptionModel *SubscriptionModel) UpdateSubscriptionCardIDToNull(uid string, cardID *string) {
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

	if _, err := subscriptionModel.Collection.UpdateMany(context.TODO(), match,
		bson.M{"$set": bson.M{
			"card_id": nil,
		}}); err != nil {
		return
	}
}

func (subscriptionModel *SubscriptionModel) CancelSubscriptionInvitation(id, uid string) error {
	objectID, _ := primitive.ObjectIDFromHex(id)

	result := subscriptionModel.InviteCollection.FindOne(context.TODO(), bson.M{"_id": objectID})

	var subscriptionInvite SubscriptionInvite
	if err := result.Decode(&subscriptionInvite); err != nil {
		return fmt.Errorf("Failed to decode invitation.")
	}

	if subscriptionInvite.UserID != uid {
		return fmt.Errorf("Unauthorized access.")
	}

	if err := subscriptionModel.UpdateSubscriptionInvite(
		subscriptionInvite.InvitedUserID, subscriptionInvite.SubscriptionID, false, false,
	); err != nil {
		return err
	}

	if _, err := subscriptionModel.InviteCollection.DeleteOne(context.TODO(), bson.M{"_id": objectID}); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Error("failed to delete invitation")

		return fmt.Errorf("Failed to delete invitation.")
	}

	return nil
}

func (subscriptionModel *SubscriptionModel) HandleSubscriptionInvitation(id, uid string, isAccepted bool) error {
	objectID, _ := primitive.ObjectIDFromHex(id)

	result := subscriptionModel.InviteCollection.FindOne(context.TODO(), bson.M{"_id": objectID})

	var subscriptionInvite SubscriptionInvite
	if err := result.Decode(&subscriptionInvite); err != nil {
		return fmt.Errorf("Failed to decode invitation.")
	}

	if subscriptionInvite.InvitedUserID != uid {
		return fmt.Errorf("Unauthorized access.")
	}

	if _, err := subscriptionModel.InviteCollection.DeleteOne(context.TODO(), bson.M{"_id": objectID}); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Error("failed to delete invitation")

		return fmt.Errorf("Failed to delete invitation.")
	}

	if err := subscriptionModel.UpdateSubscriptionInvite(subscriptionInvite.InvitedUserID, subscriptionInvite.SubscriptionID, false, isAccepted); err != nil {
		logrus.WithFields(logrus.Fields{
			"invited_uid":     subscriptionInvite.InvitedUserID,
			"subscription_id": subscriptionInvite.SubscriptionID,
		}).Error("failed to set subscription")

		return err
	}

	return nil
}

func (subscriptionModel *SubscriptionModel) DeleteSubscriptionBySubscriptionID(uid, subscriptionID string) (bool, error) {
	objectSubscriptionID, _ := primitive.ObjectIDFromHex(subscriptionID)

	count, err := subscriptionModel.Collection.DeleteOne(context.TODO(), bson.M{
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

func (subscriptionModel *SubscriptionModel) DeleteAllSubscriptionsByUserID(uid string) error {
	if _, err := subscriptionModel.Collection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all subscriptions by user id: ", err)

		return fmt.Errorf("Failed to delete all subscriptions by user id.")
	}

	return nil
}

func (subscriptionModel *SubscriptionModel) DeleteAllSubscriptionInvitesByUserID(uid string) error {
	if _, err := subscriptionModel.InviteCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all subscription invites by user id: ", err)

		return fmt.Errorf("Failed to delete all subscription invites by user id.")
	}

	return nil
}

func (subscriptionModel *SubscriptionModel) isInviteSentToUser(uid, invitedUID, subscriptionID string) bool {
	count, err := subscriptionModel.InviteCollection.CountDocuments(context.TODO(), bson.M{
		"user_id":         uid,
		"invited_user_id": invitedUID,
		"subscription_id": subscriptionID,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":             uid,
			"invited_uid":     invitedUID,
			"subscription_id": subscriptionID,
		}).Error("failed to count user invites: ", err)

		return true
	}

	return count > 0
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

	var account *responses.SubscriptionAccount

	if subscription.Account != nil {
		var decryptedPassword string
		if subscription.Account.Password != nil {
			decryptedPassword = utils.Decrypt(*subscription.Account.Password)
		}

		account = &responses.SubscriptionAccount{
			EmailAddress: subscription.Account.EmailAddress,
			Password:     &decryptedPassword,
		}
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
		BillCycle:        billCycle,
		Price:            subscription.Price,
		Currency:         subscription.Currency,
		Color:            subscription.Color,
		Image:            &subscription.Image,
		CreatedAt:        subscription.CreatedAt,
		NotificationTime: subscription.NotificationTime,
		Account:          account,
	}
}

func addSubscriptionMonthlyAndTotalPaymentFields() bson.M {
	return bson.M{"$addFields": bson.M{
		"monthly_payment": bson.M{
			"$round": bson.A{
				bson.M{
					"$switch": bson.M{
						"branches": bson.A{
							// Day case
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
							// Month Case
							bson.M{
								"case": bson.M{"$gt": bson.A{"$bill_cycle.month", 1}},
								"then": bson.M{
									"$divide": bson.A{"$price", "$bill_cycle.month"},
								},
							},
							// Year Case
							bson.M{
								"case": bson.M{"$gt": bson.A{"$bill_cycle.year", 0}},
								"then": bson.M{
									"$divide": bson.A{
										"$price",
										bson.M{
											"$multiply": bson.A{
												12,
												"$bill_cycle.year",
											},
										},
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
											// Day case
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
											// Month Case
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
											// Year Case
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
