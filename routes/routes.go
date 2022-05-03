package routes

import (
	"asset_backend/controllers"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	userController            = new(controllers.UserController)
	assetController           = new(controllers.AssetController)
	dailyAssetStatsController = new(controllers.DailyAssetStatsController)
	subscriptionController    = new(controllers.SubscriptionController)
	investingController       = new(controllers.InvestingController)
	OAuth2Controller          = new(controllers.OAuth2Controller)
	logController             = new(controllers.LogController)
)

func SetupRoutes(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware) {
	apiRouter := router.Group("/api/v1")

	userRouter(apiRouter, jwtToken)
	assetRouter(apiRouter, jwtToken)
	subscriptionRouter(apiRouter, jwtToken)
	oauth2Router(apiRouter, jwtToken)

	apiRouter.Use(jwtToken.MiddlewareFunc()).POST("/log", logController.CreateLog)
	router.GET("/confirm-password-reset", userController.ConfirmPasswordReset)
	router.GET("/privacy", privacyPolicy)
	router.GET("/terms", termsConditions)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "all routes lead to rome"})
	})
}

func privacyPolicy(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/privacy_policy.html")
}

func termsConditions(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/terms_conditions.html")
}
