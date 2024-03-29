package main

import (
	"asset_backend/controllers"
	"asset_backend/db"
	"asset_backend/docs"
	"asset_backend/helpers"
	"asset_backend/models"
	"asset_backend/routes"
	"asset_backend/utils"
	"log"
	"net/http"
	"os"
	"strconv"
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
// @termsOfService https://kanma-backend.onrender.com/terms/

// @contact.name Burak Fidan
// @contact.email mrntlu@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host https://kanma-backend.onrender.com
// @BasePath /api/v1
// @schemes https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if os.Getenv("ENV") != "Production" {
		if err := godotenv.Load(".env"); err != nil {
			log.Default().Println(os.Getenv("ENV"))
			log.Fatal("Error loading .env file")
		}
	}

	controllers.SetOAuth2()

	mongoDB, ctx, cancel := db.Connect(os.Getenv("MONGO_ATLAS_URI"))
	defer db.Close(ctx, mongoDB.Client, cancel)

	db.SetupRedis()
	utils.InitCipher()

	jwtHandler := helpers.SetupJWTHandler(mongoDB)

	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC822,
		PrettyPrint:     true,
	})

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"

	const (
		burstTime       = 100 * time.Millisecond
		requestCount    = 20
		restrictionTime = 5 * time.Second
	)
	// Burst of 0.1 sec 20 requests. 5 second restriction.
	router.Use(limit.NewRateLimiter(func(ctx *gin.Context) string {
		return ctx.ClientIP()
	}, func(ctx *gin.Context) (*rate.Limiter, time.Duration) {
		return rate.NewLimiter(rate.Every(burstTime), requestCount), restrictionTime
	}, func(ctx *gin.Context) {
		const tooManyRequestError = "Too many requests. Rescricted for 5 seconds."
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": tooManyRequestError, "message": tooManyRequestError})
		ctx.Abort()
	}))

	routes.SetupRoutes(router, jwtHandler, mongoDB)

	var dailyScheduler *gocron.Scheduler
	dailyScheduler = helpers.CreateDailySchedule(func() {
		dailyTask(mongoDB)
		scheduleLogger(dailyScheduler, "Daily")
	}, "05:00")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.Run(":" + port)
}

func dailyTask(mongoDB *db.MongoDB) {
	dasModel := models.NewDailyAssetStatsModel(mongoDB)
	go dasModel.CalculateDailyAssetStats()

	userModel := models.NewUserModel(mongoDB)
	notificationSubs := userModel.GetSubscriptionNotifications()

	for _, notificationSubscription := range notificationSubs {
		const (
			numCheck  = 10
			floatPrec = 2
			floatBit  = 64
		)

		var (
			hourString   string
			minuteString string
		)

		subscription := notificationSubscription.Subscription

		hour, min, _ := subscription.NotificationTime.Clock()

		if min < numCheck {
			minuteString = "0" + strconv.Itoa(min)
		} else {
			minuteString = strconv.Itoa(min)
		}

		if hour < numCheck {
			hourString = "0" + strconv.Itoa(hour)
		} else {
			hourString = strconv.Itoa(hour)
		}

		fcmToken := notificationSubscription.FCMToken

		helpers.CreateSubscriptionNotificationSchedule(func() {
			dataType := "subscription"
			dataID := subscription.ID.Hex()

			go helpers.SendNotification(
				fcmToken,
				subscription.Name+"'s Payment",
				"Upcoming "+subscription.Name+" Payment: "+subscription.Currency+" "+strconv.FormatFloat(subscription.Price, 'f', floatPrec, floatBit),
				&dataType,
				&dataID,
			)
		}, (hourString + ":" + minuteString))
	}
}

func scheduleLogger(scheduler *gocron.Scheduler, tType string) {
	hourlyJob, nextRun := scheduler.NextRun()

	logrus.WithFields(logrus.Fields{
		tType + " Job Last Run":       hourlyJob.LastRun(),
		tType + " Job Run Count":      hourlyJob.RunCount(),
		"Next " + tType + " Schedule": nextRun,
	}).Info("Schedule Info")
}
