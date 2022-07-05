package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"asset_backend/responses"
	"asset_backend/utils"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserModel struct {
	Collection *mongo.Collection
}

func NewUserModel(mongoDB *db.MongoDB) *UserModel {
	return &UserModel{
		Collection: mongoDB.Database.Collection("users"),
	}
}

/**
* !Premium Features
* *Asset
* 	- Max 10 asset
*	- Only weekly stats.
* *Subscription
*	- Max 5 subscription.
* *Card
*	- Max 3 cards.
* *Bank Account
*	- Max 2 bank accounts.
* *Transactions
*	- Max 10 per day.
**/
type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	EmailAddress       string             `bson:"email_address" json:"email_address"`
	Currency           string             `bson:"currency" json:"currency"`
	Password           string             `bson:"password" json:"-"`
	PasswordResetToken string             `bson:"reset_token" json:"-"`
	CreatedAt          time.Time          `bson:"created_at" json:"-"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"-"`
	IsPremium          bool               `bson:"is_premium" json:"is_premium"`
	IsLifetimePremium  bool               `bson:"is_lifetime_premium" json:"is_lifetime_premium"`
	IsOAuthUser        bool               `bson:"is_oauth" json:"is_oauth"`
	OAuthType          int                `bson:"oauth_type" json:"oauth_type"`
	RefreshToken       *string            `bson:"refresh_token" json:"-"`
	FCMToken           string             `bson:"fcm_token" json:"fcm_token"`
	AppNotification    bool               `bson:"app_notification" json:"app_notification"`
	MailNotification   bool               `bson:"mail_notification" json:"mail_notification"`
}

func createUserObject(emailAddress, currency, password string) *User {
	return &User{
		EmailAddress:      emailAddress,
		Currency:          currency,
		Password:          utils.HashPassword(password),
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
		IsPremium:         false,
		IsLifetimePremium: false,
		IsOAuthUser:       false,
		AppNotification:   true,
		MailNotification:  true,
		OAuthType:         -1,
		FCMToken:          "",
	}
}

func createOAuthUserObject(emailAddress string, refreshToken *string, oAuthType int) *User {
	return &User{
		EmailAddress:     emailAddress,
		Currency:         "USD",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
		IsPremium:        false,
		IsOAuthUser:      true,
		AppNotification:  true,
		MailNotification: false,
		OAuthType:        oAuthType,
		RefreshToken:     refreshToken,
	}
}

func (userModel *UserModel) CreateUser(data requests.Register) error {
	user := createUserObject(data.EmailAddress, data.Currency, data.Password)

	if _, err := userModel.Collection.InsertOne(context.TODO(), user); err != nil {
		logrus.WithFields(logrus.Fields{
			"email": data.EmailAddress,
		}).Error("failed to create new user: ", err)

		return fmt.Errorf("Failed to create new user.")
	}

	return nil
}

func (userModel *UserModel) CreateOAuthUser(email string, refreshToken *string, oAuthType int) (*User, error) {
	user := createOAuthUserObject(email, refreshToken, oAuthType)

	result, err := userModel.Collection.InsertOne(context.TODO(), user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"email": email,
		}).Error("failed to create new oauth user: ", err)

		return nil, fmt.Errorf("Failed to create new oauth user.")
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	return user, nil
}

func (userModel *UserModel) GetSubscriptionNotifications() []responses.NotificationSubscription {
	match := bson.M{"$match": bson.M{
		"is_premium":       true,
		"app_notification": true,
	}}
	lookup := bson.M{"$lookup": bson.M{
		"from": "subscriptions",
		"let": bson.M{
			"uid": bson.M{
				"$toString": "$_id",
			},
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"notification_time": bson.M{"$exists": true},
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$user_id", "$$uid"}},
							bson.M{"$ne": bson.A{"$notification_time", nil}},
						},
					},
				},
			},
		},
		"as": "subscriptions",
	}}
	unwindSubscription := bson.M{"$unwind": bson.M{
		"path":                       "$subscriptions",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}
	project := bson.M{"$project": bson.M{
		"subscription": "$subscriptions",
		"fcm_token":    true,
	}}

	cursor, err := userModel.Collection.Aggregate(context.TODO(), bson.A{
		match, lookup, unwindSubscription, project,
	})
	if err != nil {
		logrus.Error("failed to aggregate users on subscription notification: ", err)

		return nil
	}

	var notificationSubs []responses.NotificationSubscription
	if err = cursor.All(context.TODO(), &notificationSubs); err != nil {
		logrus.Error("failed to decode subscriptions: ", err)

		return nil
	}

	// TODO Either filter by today or send notification & update notification date

	return notificationSubs
}

func (userModel *UserModel) UpdateUser(user User) error {
	user.UpdatedAt = time.Now().UTC()

	if _, err := userModel.Collection.UpdateOne(context.TODO(), bson.M{"_id": user.ID}, bson.M{"$set": user}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": user.ID,
		}).Error("failed to update user: ", err)

		return fmt.Errorf("Failed to update user.")
	}

	return nil
}

func (userModel *UserModel) UpdateUserMembership(uid string, data requests.ChangeMembership) error {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	if _, err := userModel.Collection.UpdateOne(context.TODO(), bson.M{"_id": objectUID}, bson.M{"$set": bson.M{
		"is_premium":          data.IsPremium,
		"is_lifetime_premium": data.IsLifetimePremium,
		"updated_at":          time.Now().UTC(),
	}}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":                 uid,
			"is_premium":          data.IsPremium,
			"is_lifetime_premium": data.IsLifetimePremium,
		}).Error("failed to set membership for user: ", err)

		return fmt.Errorf("Failed to set membership for user.")
	}

	return nil
}

func (userModel *UserModel) IsUserPremium(uid string) bool {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectUID,
	})

	var isUserPremium responses.IsUserPremium
	if err := result.Decode(&isUserPremium); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find user by uid: ", err)

		return false
	}

	return isUserPremium.IsPremium || isUserPremium.IsLifetimePremium
}

func (userModel *UserModel) FindUserByID(uid string) (User, error) {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectUID,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": user.ID,
		}).Error("failed to find user by uid: ", err)

		return User{}, fmt.Errorf("Failed to find user by id.")
	}

	return user, nil
}

func (userModel *UserModel) FindUserByRefreshToken(refreshToken string) (User, error) {
	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"refresh_token": refreshToken,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"refresh_token": refreshToken,
		}).Error("failed to find user by refreshToken: ", err)

		return User{}, fmt.Errorf("Failed to find user by token.")
	}

	return user, nil
}

func (userModel *UserModel) FindUserByResetTokenAndEmail(token, email string) (User, error) {
	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"reset_token":   token,
		"email_address": email,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":   user.ID,
			"token": token,
		}).Error("failed to find user by reset token: ", err)

		return User{}, fmt.Errorf("Failed to find user by reset token.")
	}

	return user, nil
}

func (userModel *UserModel) FindUserByEmail(email string) (User, error) {
	result := userModel.Collection.FindOne(context.TODO(), bson.M{
		"email_address": email,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"email": email,
		}).Error("failed to find user by email: ", err)

		return User{}, fmt.Errorf("Failed to find user by email.")
	}

	return user, nil
}

func (userModel *UserModel) DeleteUserByID(uid string) error {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	if _, err := userModel.Collection.DeleteOne(context.TODO(), bson.M{"_id": objectUID}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete user: ", err)

		return fmt.Errorf("Failed to delete user.")
	}

	return nil
}
