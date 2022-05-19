package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func subscriptionRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware) {
	subscription := router.Group("/subscription").Use(jwtToken.MiddlewareFunc())
	{
		subscription.DELETE("/all", subscriptionController.DeleteAllSubscriptionsByUserID)
		subscription.DELETE("", subscriptionController.DeleteSubscriptionBySubscriptionID)
		subscription.PUT("", subscriptionController.UpdateSubscription)
		subscription.POST("", subscriptionController.CreateSubscription)
		subscription.GET("/card", subscriptionController.GetSubscriptionsByCardID)
		subscription.GET("", subscriptionController.GetSubscriptionsAndStatsByUserID)
		subscription.GET("/details", subscriptionController.GetSubscriptionDetails)
		subscription.GET("/stats", subscriptionController.GetSubscriptionStatisticsByUserID)
	}
}
