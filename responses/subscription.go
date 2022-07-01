package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID       string             `bson:"user_id" json:"user_id"`
	CardID       *string            `bson:"card_id" json:"card_id"`
	Name         string             `bson:"name" json:"name"`
	Description  *string            `bson:"description" json:"description"`
	BillDate     time.Time          `bson:"bill_date" json:"bill_date"`
	NextBillDate time.Time          `bson:"next_bill_date" json:"next_bill_date"`
	BillCycle    BillCycle          `bson:"bill_cycle" json:"bill_cycle"`
	Price        float64            `bson:"price" json:"price"`
	Currency     string             `bson:"currency" json:"currency"`
	Color        string             `bson:"color" json:"color"`
	Image        *string            `bson:"image" json:"image"`
	CreatedAt    time.Time          `bson:"created_at" json:"-"`
}

type SubscriptionDetails struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	UserID         string               `bson:"user_id" json:"user_id"`
	CardID         *string              `bson:"card_id" json:"card_id"`
	Name           string               `bson:"name" json:"name"`
	Description    *string              `bson:"description" json:"description"`
	BillDate       time.Time            `bson:"bill_date" json:"bill_date"`
	NextBillDate   time.Time            `bson:"next_bill_date" json:"next_bill_date"`
	BillCycle      BillCycle            `bson:"bill_cycle" json:"bill_cycle"`
	Price          float64              `bson:"price" json:"price"`
	Currency       string               `bson:"currency" json:"currency"`
	MonthlyPayment float64              `bson:"monthly_payment" json:"monthly_payment"`
	TotalPayment   float64              `bson:"total_payment" json:"total_payment"`
	Card           *Card                `bson:"card" json:"card"`
	Account        *SubscriptionAccount `json:"account"`
}

type SubscriptionAccount struct {
	EmailAddress string  `bson:"email_address" json:"email_address"`
	Password     *string `bson:"password" json:"password"`
}

type Card struct {
	Name       string `bson:"name" json:"name"`
	Last4Digit string `bson:"last_digit" json:"last_digit"`
	Type       string `bson:"type" json:"type"`
}

type BillCycle struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

type SubscriptionStatistics struct {
	ID                  string  `bson:"_id" json:"currency"`
	TotalMonthlyPayment float64 `bson:"total_monthly_payment" json:"total_monthly_payment"`
	TotalPayment        float64 `bson:"total_payment" json:"total_payment"`
}

type SubscriptionAndStats struct {
	Data  []Subscription           `bson:"data" json:"data"`
	Stats []SubscriptionStatistics `bson:"stats" json:"stats"`
}
