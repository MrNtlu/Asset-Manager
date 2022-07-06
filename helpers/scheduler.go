package helpers

import (
	"time"

	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

func CreateHourlySchedule(task interface{}, rate int) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(time.UTC)

	if _, err := scheduler.Every(rate).Hour().Do(task); err != nil {
		logrus.WithFields(logrus.Fields{
			"rate": rate,
		}).Error("error hourly schedule ", err)
	}

	scheduler.StartAsync()

	return scheduler
}

func CreateDailySchedule(task interface{}, atTime string) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(time.UTC)

	if _, err := scheduler.Every(1).Day().At(atTime).Do(task); err != nil {
		logrus.WithFields(logrus.Fields{
			"time": atTime,
		}).Error("error daily schedule ", err)
	}

	scheduler.StartAsync()

	return scheduler
}

func CreateSubscriptionNotificationSchedule(task interface{}, atTime string) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(time.UTC)

	if _, err := scheduler.Every(1).Day().At(atTime).LimitRunsTo(1).Do(task); err != nil {
		logrus.WithFields(logrus.Fields{
			"time": atTime,
		}).Error("error notification schedule ", err)
	}

	scheduler.StartAsync()

	return scheduler
}
