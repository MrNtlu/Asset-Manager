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

func GetUserTransactionCountByTime(uid string) int64 {
	//TODO: Implement
	return 10
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

//TODO: GET Transaction by uid, date && pagination
func GetTransactionsByUserID(uid string) ([]Transaction, error) {
	return nil, nil
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
