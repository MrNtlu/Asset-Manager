package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"fmt"
	"time"

	pagination "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//Categories
const (
	Food int64 = iota
	Shopping
	Transportation
	Entertainment
	Software
	Health
	Income
	Others
)

//Transaction Types
const (
	BankAcc int64 = iota
	CreditCard
)

//TODO: Allow receipt photo and keep it also get price from receipt via ML or something research (Premium feature only)
type Transaction struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID            string             `bson:"user_id" json:"user_id"`
	Title             string             `bson:"title" json:"title"`
	Description       *string            `bson:"description" json:"description"`
	Category          int64              `bson:"category" json:"category"`
	Price             float64            `bson:"price" json:"price"`
	Currency          string             `bson:"currency" json:"currency"`
	TransactionMethod *TransactionMethod `bson:"method" json:"method"`
	TransactionDate   time.Time          `bson:"transaction_date" json:"transaction_date"`
	CreatedAt         time.Time          `bson:"created_at" json:"-"`
}

type TransactionMethod struct {
	MethodID string `bson:"method_id" json:"method_id"`
	Type     int64  `bson:"type" json:"type"`
}

func createTransaction(uid, title, currency string, category int64, price float64, transactionDate time.Time, method *TransactionMethod, description *string) *Transaction {
	return &Transaction{
		UserID:            uid,
		Title:             title,
		Currency:          currency,
		Category:          category,
		Price:             price,
		TransactionDate:   transactionDate,
		TransactionMethod: method,
		Description:       description,
		CreatedAt:         time.Now().UTC(),
	}
}

func createTransactionMethod(data requests.TransactionMethod) *TransactionMethod {
	return &TransactionMethod{
		MethodID: data.MethodID,
		Type:     *data.Type,
	}
}

func CreateTransaction(uid string, data requests.TransactionCreate) (Transaction, error) {
	transaction := createTransaction(
		uid,
		data.Title,
		data.Currency,
		*data.Category,
		data.Price,
		data.TransactionDate,
		createTransactionMethod(*data.TransactionMethod),
		data.Description,
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)
	if insertedID, err = db.TransactionCollection.InsertOne(context.TODO(), transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new transaction: ", err)
		return Transaction{}, fmt.Errorf("Failed to create new transaction.")
	}
	transaction.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *transaction, nil
}

func GetTotalTransactionByInterval(uid string, data requests.TransactionTotalInterval) (responses.TransactionTotal, error) {
	var match bson.M
	if data.Interval == "day" {
		match = bson.M{"$match": bson.M{
			"user_id": uid,
			"$expr": bson.M{
				"$and": bson.A{
					bson.M{
						"$eq": bson.A{
							bson.M{"$dayOfMonth": "$transaction_date"},
							data.TransactionDate.Day(),
						},
					},
					bson.M{
						"$eq": bson.A{
							bson.M{"$month": "$transaction_date"},
							int(data.TransactionDate.Month()),
						},
					},
					bson.M{
						"$eq": bson.A{
							bson.M{"$year": "$transaction_date"},
							data.TransactionDate.Year(),
						},
					},
				},
			},
		}}
	} else {
		match = bson.M{"$match": bson.M{
			"user_id": uid,
			"$expr": bson.M{
				"$and": bson.A{
					bson.M{
						"$eq": bson.A{
							bson.M{"$month": "$transaction_date"},
							int(data.TransactionDate.Month()),
						},
					},
					bson.M{
						"$eq": bson.A{
							bson.M{"$year": "$transaction_date"},
							data.TransactionDate.Year(),
						},
					},
				},
			},
		}}
	}
	uidToObject := bson.M{"$addFields": bson.M{
		"user_id": bson.M{
			"$toObjectId": "$user_id",
		},
	}}
	userLookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "user_id",
		"foreignField": "_id",
		"as":           "user",
	}}
	unwindUser := bson.M{"$unwind": bson.M{
		"path":                       "$user",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	userCurrencyExchangeLookup := bson.M{"$lookup": bson.M{
		"from": "exchanges",
		"let": bson.M{
			"user_currency":        "$user.currency",
			"transaction_currency": "$currency",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$ne": bson.A{"$$transaction_currency", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$to_exchange", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$from_exchange", "TRY"}},
						},
					},
				},
			},
		},
		"as": "user_exchange_rate",
	}}
	unwindUserCurrency := bson.M{"$unwind": bson.M{
		"path":                       "$user_exchange_rate",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	addExhangeValue := bson.M{"$addFields": bson.M{
		"value": bson.M{
			"$ifNull": bson.A{
				bson.M{
					"$multiply": bson.A{"$price", "$user_exchange_rate.exchange_rate"},
				},
				"$price",
			},
		},
	}}
	group := bson.M{"$group": bson.M{
		"_id": "$user_id",
		"currency": bson.M{
			"$first": "$user.currency",
		},
		"total_transaction": bson.M{
			"$sum": "$value",
		},
	}}

	cursor, err := db.TransactionCollection.Aggregate(context.TODO(), bson.A{
		match, uidToObject, userLookup, unwindUser, userCurrencyExchangeLookup, unwindUserCurrency, addExhangeValue, group,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate total transactions: ", err)
		return responses.TransactionTotal{}, fmt.Errorf("Failed to aggregate total transactions.")
	}

	var totalTransaction []responses.TransactionTotal
	if err = cursor.All(context.TODO(), &totalTransaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode total transactions: ", err)
		return responses.TransactionTotal{}, fmt.Errorf("Failed to decode total transactions.")
	}

	if len(totalTransaction) > 0 {
		return totalTransaction[0], nil
	}

	return responses.TransactionTotal{}, nil
}

func GetMethodStatistics(uid string, data requests.TransactionMethod) (responses.TransactionTotal, error) {
	match := bson.M{"$match": bson.M{
		"method.type":      data.Type,
		"method.method_id": data.MethodID,
	}}
	uidToObject := bson.M{"$addFields": bson.M{
		"user_id": bson.M{
			"$toObjectId": "$user_id",
		},
	}}
	userLookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "user_id",
		"foreignField": "_id",
		"as":           "user",
	}}
	unwindUser := bson.M{"$unwind": bson.M{
		"path":                       "$user",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	userCurrencyExchangeLookup := bson.M{"$lookup": bson.M{
		"from": "exchanges",
		"let": bson.M{
			"user_currency":        "$user.currency",
			"transaction_currency": "$currency",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$ne": bson.A{"$$transaction_currency", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$to_exchange", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$from_exchange", "TRY"}},
						},
					},
				},
			},
		},
		"as": "user_exchange_rate",
	}}
	unwindUserCurrency := bson.M{"$unwind": bson.M{
		"path":                       "$user_exchange_rate",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	addExhangeValue := bson.M{"$addFields": bson.M{
		"value": bson.M{
			"$ifNull": bson.A{
				bson.M{
					"$multiply": bson.A{"$price", "$user_exchange_rate.exchange_rate"},
				},
				"$price",
			},
		},
	}}
	group := bson.M{"$group": bson.M{
		"_id": "$user_id",
		"currency": bson.M{
			"$first": "$user.currency",
		},
		"total_transaction": bson.M{
			"$sum": "$value",
		},
	}}

	cursor, err := db.TransactionCollection.Aggregate(context.TODO(), bson.A{
		match, uidToObject, userLookup, unwindUser, userCurrencyExchangeLookup, unwindUserCurrency, addExhangeValue, group,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to aggregate transactions: ", err)
		return responses.TransactionTotal{}, fmt.Errorf("Failed to aggregate transactions.")
	}

	var totalTransaction []responses.TransactionTotal
	if err = cursor.All(context.TODO(), &totalTransaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to decode transactions: ", err)
		return responses.TransactionTotal{}, fmt.Errorf("Failed to decode transactions.")
	}

	if len(totalTransaction) > 0 {
		return totalTransaction[0], nil
	}

	return responses.TransactionTotal{}, nil
}

func GetTransactionStats(uid string, data requests.TransactionStatsInterval) ([]responses.TransactionStats, error) {
	var intervalDate time.Time
	switch data.Interval {
	case "weekly":
		intervalDate = time.Now().AddDate(0, 0, -7)
	case "monthly":
		intervalDate = time.Now().AddDate(0, -1, 0)
	}

	match := bson.M{"$match": bson.M{
		"user_id": uid,
		"transaction_date": bson.M{
			"$gte": intervalDate,
		},
	}}
	addFields := bson.M{"$addFields": bson.M{
		"user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"transaction_date": bson.M{
			"$toDate": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$transaction_date",
				},
			},
		},
	}}
	userLookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "user_id",
		"foreignField": "_id",
		"as":           "user",
	}}
	unwindUser := bson.M{"$unwind": bson.M{
		"path":                       "$user",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	userCurrencyExchangeLookup := bson.M{"$lookup": bson.M{
		"from": "exchanges",
		"let": bson.M{
			"user_currency":        "$user.currency",
			"transaction_currency": "$currency",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$ne": bson.A{"$$transaction_currency", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$to_exchange", "$$user_currency"}},
							bson.M{"$eq": bson.A{"$from_exchange", "TRY"}},
						},
					},
				},
			},
		},
		"as": "user_exchange_rate",
	}}
	unwindUserCurrency := bson.M{"$unwind": bson.M{
		"path":                       "$user_exchange_rate",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}
	addExhangeValue := bson.M{"$addFields": bson.M{
		"value": bson.M{
			"$ifNull": bson.A{
				bson.M{
					"$multiply": bson.A{"$price", "$user_exchange_rate.exchange_rate"},
				},
				"$price",
			},
		},
	}}
	group := bson.M{"$group": bson.M{
		"_id": "$transaction_date",
		"currency": bson.M{
			"$first": "$user.currency",
		},
		"total_transaction": bson.M{
			"$sum": "$value",
		},
		"date": bson.M{
			"$first": "$transaction_date",
		},
	}}
	sort := bson.M{"$sort": bson.M{
		"_id": 1,
	}}

	cursor, err := db.TransactionCollection.Aggregate(context.TODO(), bson.A{
		match, addFields, userLookup, unwindUser, userCurrencyExchangeLookup, unwindUserCurrency, addExhangeValue, group, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate transaction statistics: ", err)
		return nil, fmt.Errorf("Failed to aggregate transaction statistics: %w", err)
	}

	var transactionStats []responses.TransactionStats
	if err = cursor.All(context.TODO(), &transactionStats); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode transaction statistics: ", err)
		return nil, fmt.Errorf("Failed to decode transaction statistics: %w", err)
	}

	return transactionStats, nil
}

func GetUserTransactionCountByTime(uid string, date time.Time) int64 {
	count, err := db.TransactionCollection.CountDocuments(context.TODO(), bson.M{"user_id": uid, "$expr": bson.M{
		"$and": bson.A{
			bson.M{
				"$eq": bson.A{
					bson.M{"$month": "$transaction_date"},
					int(date.Month()),
				},
			},
			bson.M{
				"$eq": bson.A{
					bson.M{"$year": "$transaction_date"},
					date.Year(),
				},
			},
			bson.M{
				"$eq": bson.A{
					bson.M{"$dayOfMonth": "$transaction_date"},
					date.Day(),
				},
			},
		},
	}})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"date": date,
		}).Error("failed to count transactions by date: ", err)
		return 10
	}

	return count
}

func GetTransactionByID(transactionID string) (Transaction, error) {
	objectTransactionID, _ := primitive.ObjectIDFromHex(transactionID)

	result := db.TransactionCollection.FindOne(context.TODO(), bson.M{"_id": objectTransactionID})

	var transaction Transaction
	if err := result.Decode(&transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"transaction_id": transaction,
		}).Error("failed to create new transaction: ", err)
		return Transaction{}, fmt.Errorf("Failed to find transaction by transaction id.")
	}

	return transaction, nil
}

func GetCalendarTransactionCount(uid string, data requests.TransactionCalendar) ([]responses.TransactionCalendarCount, error) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
		"$expr": bson.M{
			"$and": bson.A{
				bson.M{
					"$eq": bson.A{
						bson.M{"$month": "$transaction_date"},
						data.Month,
					},
				},
				bson.M{
					"$eq": bson.A{
						bson.M{"$year": "$transaction_date"},
						data.Year,
					},
				},
			},
		},
	}}
	set := bson.M{"$set": bson.M{
		"transaction_date": bson.M{
			"$dateTrunc": bson.M{
				"date": "$transaction_date",
				"unit": "day",
			},
		},
	}}
	group := bson.M{"$group": bson.M{
		"_id": "$transaction_date",
		"count": bson.M{
			"$sum": 1,
		},
	}}

	cursor, err := db.TransactionCollection.Aggregate(context.TODO(), bson.A{match, set, group})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to aggregate transactions while counting: ", err)
		return nil, fmt.Errorf("Failed to aggregate transactions while counting.")
	}

	var transactionCalendarCounts []responses.TransactionCalendarCount
	if err = cursor.All(context.TODO(), &transactionCalendarCounts); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode transaction calendar count: ", err)
		return nil, fmt.Errorf("Failed to decode transaction calendar count.")
	}

	return transactionCalendarCounts, nil
}

func GetTransactionsByUserIDAndFilterSort(uid string, data requests.TransactionSortFilter) ([]Transaction, pagination.PaginationData, error) {
	match := bson.M{}
	if data.BankAccID != nil {
		match["method.type"] = BankAcc
		match["method.method_id"] = *data.BankAccID
	} else if data.CardID != nil {
		match["method.type"] = CreditCard
		match["method.method_id"] = *data.CardID
	}

	if data.Category != nil {
		match["category"] = *data.Category
	}

	if data.StartDate != nil && data.EndDate != nil {
		match["$and"] = bson.A{
			bson.M{
				"transaction_date": bson.M{
					"$gte": data.StartDate,
				},
			},
			bson.M{
				"transaction_date": bson.M{
					"$lte": data.EndDate,
				},
			},
		}
	}

	if data.StartDate != nil && data.EndDate == nil {
		match["$expr"] = bson.M{
			"$and": bson.A{
				bson.M{
					"$eq": bson.A{
						bson.M{"$month": "$transaction_date"},
						int(data.StartDate.Month()),
					},
				},
				bson.M{
					"$eq": bson.A{
						bson.M{"$year": "$transaction_date"},
						data.StartDate.Year(),
					},
				},
				bson.M{
					"$eq": bson.A{
						bson.M{"$dayOfMonth": "$transaction_date"},
						data.StartDate.Day(),
					},
				},
			},
		}
	}

	var (
		sortType  int
		sortOrder string
	)
	if data.Sort == "date" {
		sortOrder = "transaction_date"
		sortType = data.SortType
	} else {
		sortOrder = data.Sort
		sortType = data.SortType
	}

	var transactions []Transaction
	paginatedData, err := pagination.New(db.TransactionCollection).Context(context.TODO()).
		Limit(20).Sort(sortOrder, sortType).Page(data.Page).Filter(match).Decode(&transactions).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to find transaction: ", err)
		return nil, pagination.PaginationData{}, fmt.Errorf("Failed to find transaction.")
	}

	return transactions, paginatedData.Pagination, nil
}

func UpdateTransaction(data requests.TransactionUpdate, transaction Transaction) (Transaction, error) {
	objectTransactionID, _ := primitive.ObjectIDFromHex(data.ID)

	if data.Title != nil {
		transaction.Title = *data.Title
	}

	if data.Description != nil {
		transaction.Description = data.Description
	}

	if data.Category != nil {
		transaction.Category = *data.Category
	}

	if data.Price != nil {
		transaction.Price = *data.Price
	}

	if data.Currency != nil {
		transaction.Currency = *data.Currency
	}

	if data.TransactionDate != nil {
		transaction.TransactionDate = *data.TransactionDate
	}

	if data.TransactionMethod != nil {
		transaction.TransactionMethod = createTransactionMethod(*data.TransactionMethod)
	}

	if _, err := db.TransactionCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectTransactionID,
	}, bson.M{"$set": transaction}); err != nil {
		logrus.WithFields(logrus.Fields{
			"transaction_id": data.ID,
			"data":           data,
		}).Error("failed to update transaction: ", err)
		return Transaction{}, fmt.Errorf("Failed to update transaction.")
	}

	return transaction, nil
}

func UpdateTransactionMethodIDToNull(uid string, id *string, methodType int64) {
	var match bson.M
	if id != nil {
		match = bson.M{
			"method.method_id": id,
			"method.type":      methodType,
			"user_id":          uid,
		}
	} else {
		match = bson.M{
			"user_id": uid,
		}
	}

	if _, err := db.TransactionCollection.UpdateMany(context.TODO(), match,
		bson.M{"$set": bson.M{
			"method": nil,
		}}); err != nil {
		return
	}
}

func DeleteTransactionByTransactionID(uid, transactionID string) (bool, error) {
	objectTransactionID, _ := primitive.ObjectIDFromHex(transactionID)

	count, err := db.TransactionCollection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectTransactionID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":            uid,
			"transaction_id": transactionID,
		}).Error("failed to delete transaction by transaction id: ", err)
		return false, fmt.Errorf("Failed to delete transaction by transaction id.")
	}

	return count.DeletedCount > 0, nil
}

func DeleteAllTransactionsByUserID(uid string) error {
	if _, err := db.TransactionCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all transactions by user id: ", err)
		return fmt.Errorf("Failed to delete all transactions by user id.")
	}

	return nil
}
