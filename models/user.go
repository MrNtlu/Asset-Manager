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
)

/**
* !Premium Features
* *Asset
* 	- Max 10 asset
*	- Only weekly stats
* *Subscription
*	- Max 5 subscription
* *Card
*	- Max 3 cards
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
	IsOAuthUser        bool               `bson:"is_oauth" json:"-"`
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
	}
}

func createOAuthUserObject(emailAddress string) *User {
	return &User{
		EmailAddress: emailAddress,
		Currency:     "USD",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		IsPremium:    false,
		IsOAuthUser:  true,
	}
}

func CreateUser(data requests.Register) error {
	user := createUserObject(data.EmailAddress, data.Currency, data.Password)

	if _, err := db.UserCollection.InsertOne(context.TODO(), user); err != nil {
		logrus.WithFields(logrus.Fields{
			"email": data.EmailAddress,
		}).Error("failed to create new user: ", err)
		return fmt.Errorf("failed to create new user")
	}

	return nil
}

func CreateOAuthUser(email string) (*User, error) {
	user := createOAuthUserObject(email)

	result, err := db.UserCollection.InsertOne(context.TODO(), user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"email": email,
		}).Error("failed to create new oauth user: ", err)
		return nil, fmt.Errorf("failed to create new oauth user")
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func UpdateUser(user User) error {
	user.UpdatedAt = time.Now().UTC()

	if _, err := db.UserCollection.UpdateOne(context.TODO(), bson.M{"_id": user.ID}, bson.M{"$set": user}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": user.ID,
		}).Error("failed to update user: ", err)
		return fmt.Errorf("failed to update user")
	}

	return nil
}

func UpdateUserMembership(uid string, isPremium bool) error {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	if _, err := db.UserCollection.UpdateOne(context.TODO(), bson.M{"_id": objectUID}, bson.M{"$set": bson.M{
		"is_premium": isPremium,
		"updated_at": time.Now().UTC(),
	}}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":        uid,
			"is_premium": isPremium,
		}).Error("failed to set membership for user: ", err)
		return fmt.Errorf("failed to set membership for user")
	}

	return nil
}

func IsUserPremium(uid string) bool {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	result := db.UserCollection.FindOne(context.TODO(), bson.M{
		"_id": objectUID,
	})

	var isUserPremium responses.IsUserPremium
	if err := result.Decode(&isUserPremium); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find user by uid: ", err)
		return false
	}

	return isUserPremium.IsPremium
}

func FindUserByID(uid string) (User, error) {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	result := db.UserCollection.FindOne(context.TODO(), bson.M{
		"_id": objectUID,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": user.ID,
		}).Error("failed to find user by uid: ", err)
		return User{}, fmt.Errorf("failed to find user by id")
	}

	return user, nil
}

func FindUserByResetTokenAndEmail(token, email string) (User, error) {
	result := db.UserCollection.FindOne(context.TODO(), bson.M{
		"reset_token":   token,
		"email_address": email,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":   user.ID,
			"token": token,
		}).Error("failed to find user by reset token: ", err)
		return User{}, fmt.Errorf("failed to find user by reset token")
	}

	return user, nil
}

func FindUserByEmail(email string) (User, error) {
	result := db.UserCollection.FindOne(context.TODO(), bson.M{
		"email_address": email,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		logrus.WithFields(logrus.Fields{
			"email": email,
		}).Error("failed to find user by email: ", err)
		return User{}, fmt.Errorf("failed to find user by email")
	}

	return user, nil
}

func DeleteUserByID(uid string) error {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	if _, err := db.UserCollection.DeleteOne(context.TODO(), bson.M{"_id": objectUID}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete user: ", err)
		return fmt.Errorf("failed to delete user")
	}

	return nil
}
