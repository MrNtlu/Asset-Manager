package main

import (
	"asset_backend/apis"
	"asset_backend/db"
	"asset_backend/helpers"
	"asset_backend/routes"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Running")

	if os.Getenv("ENV") != "Production" {
		if err := godotenv.Load(".env"); err != nil {
			log.Default().Println(os.Getenv("ENV"))
			log.Fatal("Error loading .env file")
		}
	}

	client, ctx, cancel, err := db.Connect(os.Getenv("MONGO_ATLAS_URI"))
	if err != nil {
		panic(err)
	}
	defer db.Close(client, ctx, cancel)

	jwtHandler := helpers.SetupJWTHandler()

	//TODO: Production
	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	routes.SetupRoutes(router, jwtHandler)

	//TODO: Change on Production
	//scheduler := helpers.CreateHourlySchedule(func() { hourlyTask() }, 1)

	//job, nextRun := scheduler.NextRun()
	//fmt.Println("\nJob Last Run: ", job.LastRun(), "\nJob Run Count: ", job.RunCount())
	//fmt.Println("Next Schedule: ", nextRun, "\n ")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	//TODO: Search new port for app engine
	router.Run(":" + port)
}

func hourlyTask() {
	go apis.GetAndCreateInvesting()
	go apis.GetExchangeRates()
}
