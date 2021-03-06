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

type TransactionModel struct {
	Collection *mongo.Collection
}

func NewTransactionModel(mongoDB *db.MongoDB) *TransactionModel {
	return &TransactionModel{
		Collection: mongoDB.Database.Collection("transactions"),
	}
}

// Categories
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

// Transaction Types
const (
	BankAcc int64 = iota
	CreditCard
)

const transactionPremiumLimit = 10

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

const transactionPaginationLimit = 20

func createTransaction(
	uid, title, currency string, category int64,
	price float64, transactionDate time.Time,
	method *TransactionMethod, description *string,
) *Transaction {
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

func (transactionModel *TransactionModel) CreateTransaction(uid string, data requests.TransactionCreate) (Transaction, error) {
	var transactionMethod *TransactionMethod
	if data.TransactionMethod != nil {
		transactionMethod = createTransactionMethod(*data.TransactionMethod)
	}

	transaction := createTransaction(
		uid,
		data.Title,
		data.Currency,
		*data.Category,
		data.Price,
		data.TransactionDate,
		transactionMethod,
		data.Description,
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = transactionModel.Collection.InsertOne(context.TODO(), transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new transaction: ", err)

		return Transaction{}, fmt.Errorf("Failed to create new transaction.")
	}

	transaction.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *transaction, nil
}

func (transactionModel *TransactionModel) GetTotalTransactionByInterval(uid string, data requests.TransactionTotalInterval) (responses.TransactionTotal, error) {
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

	cursor, err := transactionModel.Collection.Aggregate(context.TODO(), bson.A{
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

func (transactionModel *TransactionModel) GetMethodStatistics(uid string, data requests.TransactionMethod) (responses.TransactionTotal, error) {
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

	cursor, err := transactionModel.Collection.Aggregate(context.TODO(), bson.A{
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

func (transactionModel *TransactionModel) GetTransactionStats(uid string, data requests.TransactionStatsInterval) ([]responses.TransactionDailyStats, error) {
	var intervalDate time.Time

	switch data.Interval {
	case "weekly":
		intervalDate = time.Now().AddDate(0, 0, -7)
	case "monthly":
		intervalDate = time.Now().AddDate(0, -1, 0)
	}

	var match bson.M
	if data.Interval == "yearly" {
		match = bson.M{"$match": bson.M{
			"user_id": uid,
			"$expr": bson.M{
				"$eq": bson.A{
					bson.M{"$year": "$transaction_date"},
					time.Now().Year(),
				},
			},
			"category": bson.M{
				"$ne": Income,
			},
		}}
	} else {
		match = bson.M{"$match": bson.M{
			"user_id": uid,
			"transaction_date": bson.M{
				"$gte": intervalDate,
			},
			"category": bson.M{
				"$ne": Income,
			},
		}}
	}

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

	var addExhangeValue bson.M
	if data.Interval == "yearly" {
		addExhangeValue = bson.M{"$addFields": bson.M{
			"value": bson.M{
				"$ifNull": bson.A{
					bson.M{
						"$multiply": bson.A{"$price", "$user_exchange_rate.exchange_rate"},
					},
					"$price",
				},
			},
			"transaction_date": bson.M{
				"$dateTrunc": bson.M{
					"date": "$transaction_date",
					"unit": "month",
				},
			},
		}}
	} else {
		addExhangeValue = bson.M{"$addFields": bson.M{
			"value": bson.M{
				"$ifNull": bson.A{
					bson.M{
						"$multiply": bson.A{"$price", "$user_exchange_rate.exchange_rate"},
					},
					"$price",
				},
			},
		}}
	}

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

	cursor, err := transactionModel.Collection.Aggregate(context.TODO(), bson.A{
		match, addFields, userLookup, unwindUser, userCurrencyExchangeLookup, unwindUserCurrency, addExhangeValue, group, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate transaction statistics: ", err)

		return nil, fmt.Errorf("Failed to aggregate transaction statistics: %w", err)
	}

	var transactionStats []responses.TransactionDailyStats
	if err = cursor.All(context.TODO(), &transactionStats); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode transaction statistics: ", err)

		return nil, fmt.Errorf("Failed to decode transaction statistics: %w", err)
	}

	return transactionStats, nil
}

func (transactionModel *TransactionModel) GetUserTransactionCountByTime(uid string, date time.Time) int64 {
	count, err := transactionModel.Collection.CountDocuments(context.TODO(), bson.M{"user_id": uid, "$expr": bson.M{
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

		return transactionPremiumLimit
	}

	return count
}

func (transactionModel *TransactionModel) GetTransactionByID(transactionID string) (Transaction, error) {
	objectTransactionID, _ := primitive.ObjectIDFromHex(transactionID)

	result := transactionModel.Collection.FindOne(context.TODO(), bson.M{"_id": objectTransactionID})

	var transaction Transaction
	if err := result.Decode(&transaction); err != nil {
		logrus.WithFields(logrus.Fields{
			"transaction_id": transaction,
		}).Error("failed to create new transaction: ", err)

		return Transaction{}, fmt.Errorf("Failed to find transaction by transaction id.")
	}

	return transaction, nil
}

func (transactionModel *TransactionModel) GetTransactionCategoryDistribution(uid string, data requests.TransactionStatsInterval) (responses.TransactionCategoryStats, error) {
	var (
		today = time.Now()
		match bson.M
	)

	switch data.Interval {
	case "weekly":
		year, week := today.ISOWeek()
		match = bson.M{"$match": bson.M{
			"user_id": uid,
			"$expr": bson.M{
				"$and": bson.A{
					bson.M{
						"$eq": bson.A{
							bson.M{"$week": "$transaction_date"},
							week,
						},
					},
					bson.M{
						"$eq": bson.A{
							bson.M{"$year": "$transaction_date"},
							year,
						},
					},
				},
			},
		}}
	case "monthly":
		match = bson.M{"$match": bson.M{
			"user_id": uid,
			"$expr": bson.M{
				"$and": bson.A{
					bson.M{
						"$eq": bson.A{
							bson.M{"$month": "$transaction_date"},
							today.Month(),
						},
					},
					bson.M{
						"$eq": bson.A{
							bson.M{"$year": "$transaction_date"},
							today.Year(),
						},
					},
				},
			},
		}}
	case "yearly":
		match = bson.M{"$match": bson.M{
			"user_id": uid,
			"$expr": bson.M{
				"$and": bson.A{
					bson.M{
						"$eq": bson.A{
							bson.M{"$year": "$transaction_date"},
							today.Year(),
						},
					},
				},
			},
		}}
	}

	set := bson.M{"$set": bson.M{
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
	facet := bson.M{"$facet": bson.M{
		"total_transaction": bson.A{bson.M{
			"$group": bson.M{
				"_id": "$user_id",
				"total_val": bson.M{
					"$sum": "$value",
				},
			},
		}},
		"category_list": bson.A{
			bson.M{
				"$sort": bson.M{
					"category": 1,
				},
			},
			bson.M{
				"$group": bson.M{
					"_id": "$category",
					"total_transaction": bson.M{
						"$sum": "$value",
					},
				},
			},
		},
		"currency": bson.A{bson.M{
			"$project": bson.M{
				"currency": "$user.currency",
			},
		}},
	}}
	setResponse := bson.M{"$set": bson.M{
		"total_transaction": bson.M{
			"$first": "$total_transaction.total_val",
		},
		"currency": bson.M{
			"$first": "$currency.currency",
		},
	}}

	cursor, err := transactionModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, userLookup, unwindUser, userCurrencyExchangeLookup,
		unwindUserCurrency, addExhangeValue, facet, setResponse,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to aggregate transaction category stats while aggregating: ", err)

		return responses.TransactionCategoryStats{}, fmt.Errorf("Failed to aggregate transaction category stats while aggregating.")
	}

	var transactionCategoryDistribution []responses.TransactionCategoryStats
	if err = cursor.All(context.TODO(), &transactionCategoryDistribution); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode transaction category stats aggregate: ", err)

		return responses.TransactionCategoryStats{}, fmt.Errorf("Failed to decode transaction category stats aggregate.")
	}

	if len(transactionCategoryDistribution) > 0 {
		return transactionCategoryDistribution[0], nil
	}

	return responses.TransactionCategoryStats{}, nil
}

func (transactionModel *TransactionModel) GetTransactionsByUserIDAndFilterSort(
	uid string, data requests.TransactionSortFilter,
) ([]Transaction, pagination.PaginationData, error) {
	match := bson.M{}
	match["user_id"] = uid

	if data.BankAccID != nil && data.CardID != nil {
		match["$or"] = bson.A{
			bson.M{
				"method.type":      BankAcc,
				"method.method_id": data.BankAccID,
			},
			bson.M{
				"method.type":      CreditCard,
				"method.method_id": data.CardID,
			},
		}
	} else if data.BankAccID != nil {
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
					"$lte": time.Date(
						data.EndDate.Year(),
						data.EndDate.Month(),
						data.EndDate.Day(),
						23, 59, 59, 0, time.UTC,
					),
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

	paginatedData, err := pagination.New(transactionModel.Collection).Context(context.TODO()).
		Limit(transactionPaginationLimit).Sort(sortOrder, sortType).Page(data.Page).Filter(match).Decode(&transactions).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to find transaction: ", err)

		return nil, pagination.PaginationData{}, fmt.Errorf("Failed to find transaction.")
	}

	return transactions, paginatedData.Pagination, nil
}

func (transactionModel *TransactionModel) UpdateTransaction(data requests.TransactionUpdate, transaction Transaction) (Transaction, error) {
	objectTransactionID, _ := primitive.ObjectIDFromHex(data.ID)

	if data.Title != nil {
		transaction.Title = *data.Title
	}

	transaction.Description = data.Description

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

	if data.ShouldDeleteMethod != nil && *data.ShouldDeleteMethod {
		transaction.TransactionMethod = nil
	}

	if _, err := transactionModel.Collection.UpdateOne(context.TODO(), bson.M{
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

func (transactionModel *TransactionModel) UpdateTransactionMethodIDToNull(uid string, id *string, methodType int64) {
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

	if _, err := transactionModel.Collection.UpdateMany(context.TODO(), match,
		bson.M{"$set": bson.M{
			"method": nil,
		}}); err != nil {
		return
	}
}

func (transactionModel *TransactionModel) DeleteTransactionByTransactionID(uid, transactionID string) (bool, error) {
	objectTransactionID, _ := primitive.ObjectIDFromHex(transactionID)

	count, err := transactionModel.Collection.DeleteOne(context.TODO(), bson.M{
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

func (transactionModel *TransactionModel) DeleteAllTransactionsByUserID(uid string) error {
	if _, err := transactionModel.Collection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all transactions by user id: ", err)

		return fmt.Errorf("Failed to delete all transactions by user id.")
	}

	return nil
}

func (transactionModel *TransactionModel) GetTotalFromCategoryStats(stats responses.TransactionCategoryStats, isIncome bool) float64 {
	var total float64 = 0

	for _, item := range stats.CategoryList {
		if isIncome && item.CategoryID == Income {
			total = item.TotalCategoryTransaction
		} else if !isIncome && item.CategoryID != Income {
			total += item.TotalCategoryTransaction
		}
	}

	return total
}
