package utils

import (
	"asset_backend/responses"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CustomPagination struct {
	collection      *mongo.Collection
	context         context.Context
	aggregationList []bson.M
}

func Init(collection *mongo.Collection) *CustomPagination {
	return &CustomPagination{
		collection: collection,
	}
}

func (pagination *CustomPagination) Aggregation(context context.Context, aggregation []bson.M) *CustomPagination {
	pagination.context = context
	pagination.aggregationList = aggregation

	return pagination
}

func (pagination *CustomPagination) KeysetPaginate(lastItem string, limit int64) *CustomPagination {
	sort := bson.M{"$sort": bson.M{
		"symbol": 1,
	}}
	match := bson.M{"$match": bson.M{
		"symbol": bson.M{
			"$gt": lastItem,
		},
	}}
	limitAgg := bson.M{"$limit": limit}

	paginationAggregation := []bson.M{sort, match, limitAgg}
	pagination.aggregationList = append(pagination.aggregationList, paginationAggregation...)

	return pagination
}

func (pagination *CustomPagination) SkipLimitPaginate(limit, page int64) *CustomPagination {
	skip := bson.M{"$skip": (page - 1) * limit}
	limitAgg := bson.M{"$limit": limit}

	paginationAggregation := []bson.M{skip, limitAgg}
	pagination.aggregationList = append(pagination.aggregationList, paginationAggregation...)

	return pagination
}

func (pagination *CustomPagination) SkipLimitDecode() (*mongo.Cursor, responses.PaginationResponse, error) {
	cursor, err := pagination.collection.Aggregate(pagination.context, pagination.aggregationList)
	if err != nil {
		return nil, responses.PaginationResponse{}, fmt.Errorf("failed to aggregate pagination: %w", err)
	}

	return cursor, responses.PaginationResponse{}, nil
}

func (pagination *CustomPagination) Paginate(limit, page int64) *CustomPagination {
	facet := bson.M{"$facet": bson.M{
		"data_info": bson.A{
			bson.M{
				"$count": "total",
			},
			bson.M{
				"$addFields": bson.M{"page": page},
			},
		},
		"metadata": bson.A{
			bson.M{"$skip": (page - 1) * limit},
			bson.M{"$limit": limit},
			bson.M{
				"$group": bson.M{
					"_id":   nil,
					"count": bson.M{"$sum": 1},
					"data":  bson.M{"$push": "$$ROOT"},
				},
			},
		},
	}}
	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$metadata",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}
	project := bson.M{"$project": bson.M{
		"data": "$metadata.data",
		"metadata": bson.M{
			"$mergeObjects": bson.A{
				bson.M{
					"_id":   "$metadata._id",
					"count": "$metadata.count",
				},
				bson.M{
					"$arrayElemAt": bson.A{"$data_info", 0},
				},
			},
		},
	}}
	paginationAggregation := []bson.M{facet, unwind, project}
	pagination.aggregationList = append(pagination.aggregationList, paginationAggregation...)

	return pagination
}

func (pagination *CustomPagination) Decode() (*mongo.Cursor, responses.PaginationResponse, error) {
	cursor, err := pagination.collection.Aggregate(pagination.context, pagination.aggregationList)
	if err != nil {
		return nil, responses.PaginationResponse{}, fmt.Errorf("failed to aggregate pagination: %w", err)
	}

	defer cursor.Close(pagination.context)

	var paginationData responses.PaginationResponse
	for cursor.Next(pagination.context) {
		if err = cursor.Decode(&paginationData); err != nil {
			return nil, responses.PaginationResponse{}, fmt.Errorf("failed to aggregate decode: %w", err)
		}
	}

	return cursor, paginationData, nil
}
