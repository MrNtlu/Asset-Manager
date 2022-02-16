package routes

import (
	"asset_backend/controllers"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	userController         = new(controllers.UserController)
	assetController        = new(controllers.AssetController)
	subscriptionController = new(controllers.SubscriptionController)
	investingController    = new(controllers.InvestingController)
)

func SetupRoutes(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware) {
	userRouter(router, jwtToken)
	assetRouter(router, jwtToken)
	subscriptionRouter(router, jwtToken)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "all routes lead to rome"})
	})
}
