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
	OAuth2Controller       = new(controllers.OAuth2Controller)
)

func SetupRoutes(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware) {
	userRouter(router, jwtToken)
	assetRouter(router, jwtToken)
	subscriptionRouter(router, jwtToken)
	oauth2Router(router, jwtToken)

	router.GET("/privacy", privacyPolicy)
	router.GET("/terms", termsConditions)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "all routes lead to rome"})
	})
}

//TODO: https://search.google.com/search-console/welcome?utm_source=wmx&utm_medium=deprecation-pane&utm_content=home
// Verify page ownership
func privacyPolicy(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/privacy_policy.html")
}

func termsConditions(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/terms_conditions.html")
}
