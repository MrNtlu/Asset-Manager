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

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	client, ctx, cancel, err := db.Connect(os.Getenv("MONGO_URI"))
	if err != nil {
		panic(err)
	}
	defer db.Close(client, ctx, cancel)

	jwtHandler := helpers.SetupJWTHandler()

	router := gin.Default()
	routes.SetupRoutes(router, jwtHandler)

	scheduler := helpers.CreateHourlySchedule(func() { hourlyTask() }, 1)
	stockScheduler := helpers.CreateHourlySchedule(func() { stockTask() }, 8)

	fmt.Println(scheduler.NextRun())
	fmt.Println(stockScheduler.NextRun())

	router.Run(":8080")
}

func hourlyTask() {
	apis.GetCryptocurrencyPrices()
	apis.GetExchangeRates()
}

func stockTask() {

}
