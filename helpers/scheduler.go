package helpers

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
)

func CreateHourlySchedule(task interface{}, rate int) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(time.Local)

	if _, err := scheduler.Every(rate).Hour().Do(task); err != nil {
		fmt.Println("error schedule %w", err)
	}
	scheduler.StartAsync()

	return scheduler
}
