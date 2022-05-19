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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BankAccount struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	Name          string             `bson:"name" json:"name"`
	Iban          string             `bson:"iban" json:"iban"`
	AccountHolder string             `bson:"account_holder" json:"account_holder"`
	Currency      string             `bson:"currency" json:"currency"`
	CreatedAt     time.Time          `bson:"created_at" json:"-"`
}

/*TODO:
Stats:
	Total paid this month (transactions)
	Total expenses
*/

func createBankAccount(uid, name, iban, accoutHolder, currency string) *BankAccount {
	return &BankAccount{
		UserID:        uid,
		Name:          name,
		Iban:          iban,
		AccountHolder: accoutHolder,
		Currency:      currency,
		CreatedAt:     time.Now().UTC(),
	}
}

func CreateBankAccount(uid string, data requests.BankAccountCreate) (BankAccount, error) {
	bankAccount := createBankAccount(uid, data.Name, data.Iban, data.AccountHolder, data.Currency)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)
	if insertedID, err = db.BankAccountCollection.InsertOne(context.TODO(), bankAccount); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new bank account: ", err)
		return BankAccount{}, fmt.Errorf("Failed to create new bank account.")
	}
	bankAccount.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *bankAccount, nil
}

func GetUserBankAccountCount(uid string) int64 {
	count, err := db.BankAccountCollection.CountDocuments(context.TODO(), bson.M{"user_id": uid})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to count bank accounts: ", err)
		return 2
	}

	return count
}

func GetBankAccountByID(baID string) (BankAccount, error) {
	objectBankAccountID, _ := primitive.ObjectIDFromHex(baID)

	result := db.BankAccountCollection.FindOne(context.TODO(), bson.M{"_id": objectBankAccountID})

	var bankAccount BankAccount
	if err := result.Decode(&bankAccount); err != nil {
		logrus.WithFields(logrus.Fields{
			"bankaccount_id": baID,
		}).Error("failed to find bank account by ba id: ", err)
		return BankAccount{}, fmt.Errorf("Failed to find bank account by id.")
	}

	return bankAccount, nil
}

func GetBankAccountsByUserID(uid string) ([]BankAccount, error) {
	match := bson.M{
		"user_id": uid,
	}
	sort := bson.M{
		"created_at": 1,
	}
	options := options.Find().SetSort(sort)

	cursor, err := db.BankAccountCollection.Find(context.TODO(), match, options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find bank account by user id: ", err)
		return nil, fmt.Errorf("Failed to find bank account by user id.")
	}

	var bankAccounts []BankAccount
	if err := cursor.All(context.TODO(), &bankAccounts); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode bankAccount: ", err)
		return nil, fmt.Errorf("Failed to decode bankAccount.")
	}

	return bankAccounts, nil
}

func UpdateBankAccount(data requests.BankAccountUpdate, bankAccount BankAccount) (BankAccount, error) {
	objectBankAccountID, _ := primitive.ObjectIDFromHex(data.ID)

	if data.AccountHolder != nil {
		bankAccount.AccountHolder = *data.AccountHolder
	}

	if data.Currency != nil {
		bankAccount.Currency = *data.Currency
	}

	if data.Name != nil {
		bankAccount.Name = *data.Name
	}

	if data.Iban != nil {
		bankAccount.Iban = *data.Iban
	}

	if _, err := db.BankAccountCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectBankAccountID,
	}, bson.M{"$set": bankAccount}); err != nil {
		logrus.WithFields(logrus.Fields{
			"bankaccount_id": data.ID,
			"data":           data,
		}).Error("failed to update bank account: ", err)
		return BankAccount{}, fmt.Errorf("Failed to update bank account.")
	}

	return bankAccount, nil
}

func DeleteBankAccountByBAID(uid, baID string) (bool, error) {
	objectBankAccountID, _ := primitive.ObjectIDFromHex(baID)

	count, err := db.BankAccountCollection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectBankAccountID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":            uid,
			"bankaccount_id": baID,
		}).Error("failed to delete bank account by ba id: ", err)
		return false, fmt.Errorf("Failed to delete bank account by bank account id.")
	}

	return count.DeletedCount > 0, nil
}

func DeleteAllBankAccountsByUserID(uid string) error {
	if _, err := db.BankAccountCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all bank accounts by user id: ", err)
		return fmt.Errorf("Failed to delete all bank accounts by user id.")
	}

	return nil
}
