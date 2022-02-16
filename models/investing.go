package models

import (
	"asset_backend/db"
	"asset_backend/responses"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func GetInvestingsByType(tType string) ([]responses.InvestingResponse, error) {
	match := bson.M{"$match": bson.M{
		"_id.type": tType,
	}}
	project := bson.M{"$project": bson.M{
		"name":   "$name",
		"symbol": "$_id.symbol",
	}}

	cursor, err := db.InvestingCollections.Aggregate(context.TODO(), bson.A{match, project})
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate investings: %w", err)
	}

	var investings []responses.InvestingResponse
	if err = cursor.All(context.TODO(), &investings); err != nil {
		return nil, fmt.Errorf("failed to decode investings: %w", err)
	}

	return investings, nil
}
