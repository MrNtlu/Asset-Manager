package helpers

import (
	"os"

	"github.com/appleboy/go-fcm"
	"github.com/sirupsen/logrus"
)

func SendNotification(deviceToken, title, message string, dataType, dataID *string) error {
	var notification *fcm.Message

	if dataType != nil && dataID != nil {
		notification = &fcm.Message{
			To: deviceToken,
			Data: map[string]interface{}{
				"type": &dataType,
				"id":   &dataID,
			},
			Notification: &fcm.Notification{
				Title: title,
				Body:  message,
				Badge: "1",
			},
		}
	} else {
		notification = &fcm.Message{
			To:   deviceToken,
			Data: nil,
			Notification: &fcm.Notification{
				Title: title,
				Body:  message,
				Badge: "1",
			},
		}
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
