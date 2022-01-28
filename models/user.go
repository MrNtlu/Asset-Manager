package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"asset_backend/utils"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	EmailAddress       string             `bson:"email_address" json:"email_address"`
	Currency           string             `bson:"currency" json:"currency"`
	Password           string             `bson:"password" json:"-"`
	PasswordResetToken string             `bson:"reset_token" json:"-"`
	CreatedAt          time.Time          `bson:"created_at" json:"-"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"-"`
}

func createUserObject(emailAddress, currency, password string) *User {
	return &User{
		EmailAddress: emailAddress,
		Currency:     currency,
		Password:     utils.HashPassword(password),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

func CreateUser(data requests.Register) error {
	user := createUserObject(data.EmailAddress, data.Currency, data.Password)

	if _, err := db.UserCollection.InsertOne(context.TODO(), user); err != nil {
		return fmt.Errorf("failed to create new user: %w", err)
	}

	return nil
}

func UpdateUser(user User) error {
	user.UpdatedAt = time.Now().UTC()

	if _, err := db.UserCollection.UpdateOne(context.TODO(), bson.M{"_id": user.ID}, bson.M{"$set": user}); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func FindUserByID(uid string) (User, error) {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	result := db.UserCollection.FindOne(context.TODO(), bson.M{
		"_id": objectUID,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		return User{}, fmt.Errorf("failed to find user by uid: %w", err)
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
		return User{}, fmt.Errorf("failed to find user by reset token: %w", err)
	}

	return user, nil
}

func FindUserByEmail(email string) (User, error) {
	result := db.UserCollection.FindOne(context.TODO(), bson.M{
		"email_address": email,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		return User{}, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}

func DeleteUserByID(uid string) error {
	objectUID, _ := primitive.ObjectIDFromHex(uid)

	if _, err := db.UserCollection.DeleteOne(context.TODO(), bson.M{"_id": objectUID}); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
