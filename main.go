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
	/*scheduler := helpers.CreateHourlySchedule(func() { hourlyTask() }, 1)

	job, nextRun := scheduler.NextRun()
	fmt.Println("\nJob Last Run: ", job.LastRun(), "\nJob Run Count: ", job.RunCount())
	fmt.Println("Next Schedule: ", nextRun, "\n ")
	*/

	//TODO: Search new port for app engine
	println("Port is ", os.Getenv("PORT"))
	router.Run(":8080")
}

func hourlyTask() {
	apis.GetAndCreateInvesting()
}
