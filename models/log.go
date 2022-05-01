package models

import (
	"asset_backend/db"
	"asset_backend/requests"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Log struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"-"`
	Log       string             `bson:"log" json:"log"`
	LogType   int                `bson:"log_type" json:"log_type"`
}

const (
	Error    = 0
	Purchase = 1
	Other    = 2
)

func createLogObject(uid, log string, logType int) *Log {
	return &Log{
		UserID:    uid,
		Log:       log,
		LogType:   logType,
		CreatedAt: time.Now().UTC(),
	}
}

func CreateLog(uid string, data requests.CreateLog) {
	log := createLogObject(uid, data.Log, data.LogType)

	if _, err := db.LogCollection.InsertOne(context.TODO(), log); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to create new log: ", err)
	}
}

func DeleteAllLogsByUserID(uid string) {
	if _, err := db.LogCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all logs by user id: ", err)
	}
}
