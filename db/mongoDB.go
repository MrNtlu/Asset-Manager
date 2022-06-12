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
	BankAccountCollection    *mongo.Collection
	TransactionCollection    *mongo.Collection
	UserCollection           *mongo.Collection
	InvestingCollection      *mongo.Collection
	ExchangeCollection       *mongo.Collection
	DailyAssetStatCollection *mongo.Collection
	LogCollection            *mongo.Collection
)

func Close(ctx context.Context, client *mongo.Client, cancel context.CancelFunc) {
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func Connect(uri string) (*mongo.Client, context.Context, context.CancelFunc) {
	const timeOut = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	Database = client.Database("asset-manager")
	AssetCollection = Database.Collection("assets")
	SubscriptionCollection = Database.Collection("subscriptions")
	CardCollection = Database.Collection("cards")
	BankAccountCollection = Database.Collection("bank-accounts")
	TransactionCollection = Database.Collection("transactions")
	UserCollection = Database.Collection("users")
	InvestingCollection = Database.Collection("investings")
	ExchangeCollection = Database.Collection("exchanges")
	DailyAssetStatCollection = Database.Collection("daily-asset-stats")
	LogCollection = Database.Collection("logs")

	return client, ctx, cancel
}
