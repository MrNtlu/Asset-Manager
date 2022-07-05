package requests

import "time"

type Subscription struct {
	CardID           *string              `json:"card_id"`
	Name             string               `json:"name" binding:"required"`
	Description      *string              `json:"description"`
	BillDate         time.Time            `json:"bill_date" binding:"required" time_format:"2006-01-02"`
	BillCycle        BillCycle            `json:"bill_cycle" binding:"required"`
	Price            float64              `json:"price" binding:"required"`
	Currency         string               `json:"currency" binding:"required"`
	Color            string               `json:"color" binding:"required"`
	Image            string               `json:"image"`
	Account          *SubscriptionAccount `json:"account"`
	NotificationTime *time.Time           `json:"notification_time"`
}

type SubscriptionAccount struct {
	EmailAddress string  `bson:"email_address" json:"email_address"`
	Password     *string `bson:"password" json:"password"`
}

type BillCycle struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

type SubscriptionUpdate struct {
	ID               string               `json:"id" binding:"required"`
	Name             *string              `json:"name"`
	Description      *string              `json:"description"`
	BillDate         *time.Time           `json:"bill_date"`
	BillCycle        *BillCycle           `json:"bill_cycle"`
	Price            *float64             `json:"price"`
	Currency         *string              `json:"currency"`
	CardID           *string              `json:"card_id"`
	Color            *string              `json:"color"`
	Image            *string              `json:"image"`
	Account          *SubscriptionAccount `json:"account"`
	NotificationTime *time.Time           `json:"notification_time" time_format:"2006-01-02"`
}

type SubscriptionSort struct {
	Sort     string `form:"sort" binding:"required,oneof=name currency price date"`
	SortType int    `form:"type" json:"type" binding:"required,oneof=1 -1"`
}

type SubscriptionInvite struct {
	ID              string `json:"id" binding:"required"`
	InvitedUserMail string `json:"invited_user_mail" binding:"required"`
}

type SubscriptionInvitation struct {
	ID         string `json:"id" binding:"required"`
	IsAccepted *bool  `json:"is_accepted" binding:"required"`
}
