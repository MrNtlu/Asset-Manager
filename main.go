package main

import (
	"asset_backend/apis"
	"asset_backend/controllers"
	"asset_backend/db"
	"asset_backend/helpers"
	"asset_backend/routes"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
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

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	routes.SetupRoutes(router, jwtHandler)

	scheduler := helpers.CreateHourlySchedule(func() { hourlyTask() }, 1)

	job, nextRun := scheduler.NextRun()
	logrus.WithFields(logrus.Fields{
		"Job Last Run":  job.LastRun(),
		"Job Run Count": job.RunCount(),
		"Next Schedule": nextRun,
	}).Info("Schedule Info")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}

func hourlyTask() {
	go apis.GetAndCreateInvesting()
}
