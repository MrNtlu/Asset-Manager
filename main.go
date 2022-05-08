package main

import (
	"asset_backend/controllers"
	"asset_backend/db"
	"asset_backend/docs"
	"asset_backend/helpers"
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/routes"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	limit "github.com/yangxikun/gin-limit-by-key"
	"golang.org/x/time/rate"
)

// @title Kantan Investment Manager API
// @version 1.0
// @description REST Api of Kantan.
// @termsOfService  https://rocky-reaches-65250.herokuapp.com/terms/

// @contact.name Burak Fidan
// @contact.email mrntlu@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host rocky-reaches-65250.herokuapp.com
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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

	db.SetupRedis()

	jwtHandler := helpers.SetupJWTHandler()

	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC822,
		PrettyPrint:     true,
	})

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"

	// Burst of 0.1 sec 20 requests. 5 second restriction.
	router.Use(limit.NewRateLimiter(func(ctx *gin.Context) string {
		return ctx.ClientIP()
	}, func(ctx *gin.Context) (*rate.Limiter, time.Duration) {
		return rate.NewLimiter(rate.Every(100*time.Millisecond), 20), 5 * time.Second
	}, func(ctx *gin.Context) {
		go models.CreateLog(ctx.ClientIP(), requests.CreateLog{
			Log:     "Rate-Limit",
			LogType: 0,
		})
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Rescricted for 5 seconds.", "message": "Too many requests. Rescricted for 5 seconds."})
		ctx.Abort()
	}))

	routes.SetupRoutes(router, jwtHandler)

	dailyScheduler := helpers.CreateDailySchedule(func() { dailyTask() }, "05:00")
	scheduleLogger(dailyScheduler, "Daily")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.Run(":" + port)
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
