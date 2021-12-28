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

// var userCollection = db.Database.Collection("users")

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Username     string             `bson:"username" json:"username"`
	EmailAddress string             `bson:"email_address" json:"email_address"`
	Password     string             `bson:"password" json:"-"`
	CreatedAt    time.Time          `bson:"created_at" json:"-"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"-"`
}

func createUserObject(username, emailAddress, password string) *User {
	return &User{
		Username:     username,
		EmailAddress: emailAddress,
		Password:     utils.HashPassword(password),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

func CreateUser(data requests.Register) error {
	user := createUserObject(data.Username, data.EmailAddress, data.Password)

	if _, err := db.UserCollection.InsertOne(context.TODO(), user); err != nil {
		return fmt.Errorf("failed to create new user: %w", err)
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
		return User{}, fmt.Errorf("failed to find new user by uid: %w", err)
	}

	return user, nil
}

func FindUserByEmail(email string) (User, error) {
	result := db.UserCollection.FindOne(context.TODO(), bson.M{
		"email_address": email,
	})

	var user User
	if err := result.Decode(&user); err != nil {
		return User{}, fmt.Errorf("failed to find new user by email: %w", err)
	}

	return user, nil
}
