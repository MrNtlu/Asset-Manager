package routes

import (
	"asset_backend/db"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	apiRouter := router.Group("/api/v1")

	userRouter(apiRouter, jwtToken, mongoDB)
	assetRouter(apiRouter, jwtToken, mongoDB)
	subscriptionRouter(apiRouter, jwtToken, mongoDB)
	cardRouter(apiRouter, jwtToken, mongoDB)
	bankAccountRouter(apiRouter, jwtToken, mongoDB)
	transactionRouter(apiRouter, jwtToken, mongoDB)
	oauth2Router(apiRouter, jwtToken, mongoDB)
	logRouter(apiRouter, jwtToken, mongoDB)
	favouriteInvestingRouter(apiRouter, jwtToken, mongoDB)

	router.GET("/privacy", privacyPolicy)
	router.GET("/terms", termsConditions)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "All routes lead to rome"})
	})
}

func privacyPolicy(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/privacy_policy.html")
}

func termsConditions(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/terms_conditions.html")
}
