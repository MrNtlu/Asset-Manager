package main

import (
	"asset_backend/db"
	"asset_backend/helpers"
	"asset_backend/routes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var billDate validator.Func = func(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if ok {
		today := time.Now()
		if today.After(date) {
			return false
		}
	}
	return true
}

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
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("billDate", billDate)
	}

	routes.SetupRoutes(router, jwtHandler)
	router.Run(":8080")
}
