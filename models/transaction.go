package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ( //Categories
	Food int16 = iota
	Shopping
	Transportation
	Entertainment
	Others
	//TODO: Add categories
)

//TODO: Allow receipt photo and keep it also get price from receipt via ML or something research (Premium feature only)
type Transaction struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	Title           string             `bson:"title" json:"title"`
	Description     *string            `bson:"description" json:"description"`
	Category        int16              `bson:"category" json:"category"`
	Price           float64            `bson:"price" json:"price"`
	Currency        string             `bson:"currency" json:"currency"`
	TransactionDate time.Time          `bson:"transaction_date" json:"transaction_date"`
	CreatedAt       time.Time          `bson:"created_at" json:"-"`
}
