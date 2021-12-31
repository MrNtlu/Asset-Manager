package main

import (
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
	router.Run(":8080")
}
