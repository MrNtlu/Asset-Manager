package main

import (
	"asset_backend/apis"
	"asset_backend/controllers"
	"asset_backend/db"
	"asset_backend/helpers"
	"asset_backend/models"
	"asset_backend/routes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("Running")

	if os.Getenv("ENV") != "Production" {
		if err := godotenv.Load(".env"); err != nil {
			log.Default().Println(os.Getenv("ENV"))
			log.Fatal("Error loading .env file")
		}
	}

	controllers.SetOAuth2()

	client, ctx, cancel, err := db.Connect(os.Getenv("MONGO_ATLAS_URI"))
	if err != nil {
		panic(err)
	}
	defer db.Close(client, ctx, cancel)

	jwtHandler := helpers.SetupJWTHandler()

	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC822,
		PrettyPrint:     true,
	})

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	routes.SetupRoutes(router, jwtHandler)

	hourlyScheduler := helpers.CreateHourlySchedule(func() { hourlyTask() }, 1)
	dailyScheduler := helpers.CreateDailySchedule(func() { dailyTask() }, "05:00")

	scheduleLogger(hourlyScheduler, "Hourly")
	scheduleLogger(dailyScheduler, "Daily")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}

func hourlyTask() {
	go apis.GetAndCreateInvesting()
}

func dailyTask() {
	go models.CalculateDailyAssetStats()
}

func scheduleLogger(scheduler *gocron.Scheduler, tType string) {
	hourlyJob, nextRun := scheduler.NextRun()

	logrus.WithFields(logrus.Fields{
		tType + " Job Last Run":       hourlyJob.LastRun(),
		tType + " Job Run Count":      hourlyJob.RunCount(),
		"Next " + tType + " Schedule": nextRun,
	}).Info("Schedule Info")
}
