package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DailyAssetStatsCalculation struct {
	Currency    string             `bson:"currency" json:"currency"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	TotalAssets float64            `bson:"total_assets" json:"total_assets"`
	TotalPL     float64            `bson:"total_p/l" json:"total_p/l"`
}

type DailyAssetStats struct {
	Currency    string    `bson:"currency" json:"currency"`
	TotalAssets []float64 `bson:"total_assets" json:"total_assets"`
	TotalPL     []float64 `bson:"total_p/l" json:"total_p/l"`
}
