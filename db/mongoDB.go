package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Database                 *mongo.Database
	AssetCollection          *mongo.Collection
	SubscriptionCollection   *mongo.Collection
	CardCollection           *mongo.Collection
	UserCollection           *mongo.Collection
	InvestingCollection      *mongo.Collection
	ExchangeCollection       *mongo.Collection
	DailyAssetStatCollection *mongo.Collection
	LogCollection            *mongo.Collection
)

func Close(client *mongo.Client, ctx context.Context,
	cancel context.CancelFunc) {

	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func Connect(uri string) (*mongo.Client, context.Context,
	context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	Database = client.Database("asset-manager")
	AssetCollection = Database.Collection("assets")
	SubscriptionCollection = Database.Collection("subscriptions")
	CardCollection = Database.Collection("cards")
	UserCollection = Database.Collection("users")
	InvestingCollection = Database.Collection("investings")
	ExchangeCollection = Database.Collection("exchanges")
	DailyAssetStatCollection = Database.Collection("daily-asset-stats")
	LogCollection = Database.Collection("logs")

	return client, ctx, cancel, err
}
