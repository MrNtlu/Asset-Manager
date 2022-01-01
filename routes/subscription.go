package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func subscriptionRouter(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware) {
	subscription := router.Group("/subscription")
	{
		subscription.DELETE("/all", subscriptionController.DeleteAllSubscriptionsByUserID)
		subscription.DELETE("", subscriptionController.DeleteSubscriptionBySubscriptionID)
		subscription.PUT("", subscriptionController.UpdateSubscription)
		subscription.POST("", subscriptionController.CreateSubscription)
		subscription.GET("/card", subscriptionController.GetSubscriptionsByCardID)
		subscription.GET("", subscriptionController.GetSubscriptionsByUserID)
		subscription.GET("/details", subscriptionController.GetSubscriptionDetails)

		subscription.Use(jwtToken.MiddlewareFunc())
		{
			//subscription.POST("", subscriptionController.CreateSubscription)
		}
	}

	card := router.Group("/card")
	{
		card.DELETE("/all", subscriptionController.DeleteAllCardsByUserID)
		card.DELETE("", subscriptionController.DeleteCardByCardID)
		card.PUT("", subscriptionController.UpdateCard)
		card.POST("", subscriptionController.CreateCard)
		card.GET("", subscriptionController.GetCardsByUserID)

		card.Use(jwtToken.MiddlewareFunc())
		{
			//card.POST("", subscriptionController.CreateCard)
		}
	}
}
