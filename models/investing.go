package models

import (
	"asset_backend/db"
	"asset_backend/responses"
	"asset_backend/utils"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func GetInvestingsByType(tType string) (responses.InvestingListResponse, error) {
	match := bson.M{"$match": bson.M{
		"_id.type": tType,
	}}
	project := bson.M{"$project": bson.M{
		"name":   "$name",
		"symbol": "$_id.symbol",
	}}

	// cursor, pagination, err := utils.Init(db.InvestingCollections).Aggregation(context.TODO(), []bson.M{match, project}).
	// 	Paginate(100, 380).Decode()
	// if err != nil {
	// 	return responses.InvestingListResponse{}, fmt.Errorf("%w", err)
	// }

	// var investings responses.InvestingListResponse
	// if err = cursor.Decode(&investings); err != nil {
	// 	return responses.InvestingListResponse{}, fmt.Errorf("failed to decode investings %w", err)
	// }
	// investings.Pagination = pagination.Metadata

	// return investings, nil

	// cursor, pagination, err := utils.Init(db.InvestingCollections).Aggregation(context.TODO(), []bson.M{match, project}).
	// 	SkipLimitPaginate(100, 380).SkipLimitDecode()
	// if err != nil {
	// 	return responses.InvestingListResponse{}, fmt.Errorf("%w", err)
	// }

	// var investings []responses.InvestingResponse
	// if err = cursor.All(context.TODO(), &investings); err != nil {
	// 	return responses.InvestingListResponse{}, fmt.Errorf("failed to decode investings %w", err)
	// }

	// return responses.InvestingListResponse{
	// 	Data:       investings,
	// 	Pagination: pagination.Metadata,
	// }, nil

	cursor, pagination, err := utils.Init(db.InvestingCollections).Aggregation(context.TODO(), []bson.M{match, project}).
		KeysetPaginate("O", 380).SkipLimitDecode()
	if err != nil {
		return responses.InvestingListResponse{}, fmt.Errorf("%w", err)
	}

	var investings []responses.InvestingResponse
	if err = cursor.All(context.TODO(), &investings); err != nil {
		return responses.InvestingListResponse{}, fmt.Errorf("failed to decode investings %w", err)
	}

	return responses.InvestingListResponse{
		Data:       investings,
		Pagination: pagination.Metadata,
	}, nil
}

/*
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
*/
