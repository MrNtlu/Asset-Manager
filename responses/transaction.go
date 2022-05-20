package responses

import "time"

type TransactionCalendarCount struct {
	ID    time.Time `bson:"_id" json:"_id"`
	Count int       `bson:"count" json:"count"`
}
