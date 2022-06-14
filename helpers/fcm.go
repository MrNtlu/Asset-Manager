package helpers

import (
	"os"

	"github.com/appleboy/go-fcm"
	"github.com/sirupsen/logrus"
)

func SendNotification(deviceToken, title, message string) error {
	notification := &fcm.Message{
		To: deviceToken,
		Data: map[string]interface{}{
			"type": "subscription",
			"id":   "test",
		},
		Notification: &fcm.Notification{
			Title: title,
			Body:  message,
			Badge: "1",
		},
	}

	client, err := fcm.NewClient(os.Getenv("FCM_KEY"))
	if err != nil {
		logrus.Error(err.Error(), err)
	}

	response, err := client.Send(notification)
	if err != nil {
		logrus.Error(err.Error(), err)
	}

	return response.Error
}
