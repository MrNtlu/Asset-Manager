package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
