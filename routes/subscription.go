package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func subscriptionRouter(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware) {
	subscription := router.Group("/subscription")
	{
		subscription.POST("", subscriptionController.CreateSubscription)
	}

	card := router.Group("/card")
	{
		card.POST("", subscriptionController.CreateCard)
	}
}
