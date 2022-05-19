package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func transactionRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware) {
	transaction := router.Group("/transaction").Use(jwtToken.MiddlewareFunc())
	{
		transaction.DELETE("/all", transactionController.DeleteAllTransactionsByUserID)
		transaction.DELETE("", transactionController.DeleteTransactionByTransactionID)
		transaction.POST("", transactionController.CreateTransaction)
		transaction.PUT("", transactionController.UpdateTransaction)
		transaction.GET("", transactionController.GetTransactionsByUserID)
	}
}
