package routes

import (
	"asset_backend/controllers"
	"asset_backend/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func transactionRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	transactionController := controllers.NewTransactionController(mongoDB)

	transaction := router.Group("/transaction").Use(jwtToken.MiddlewareFunc())
	{
		transaction.DELETE("/all", transactionController.DeleteAllTransactionsByUserID)
		transaction.DELETE("", transactionController.DeleteTransactionByTransactionID)
		transaction.POST("", transactionController.CreateTransaction)
		transaction.PUT("", transactionController.UpdateTransaction)
		transaction.GET("", transactionController.GetTransactionsByUserIDAndFilterSort)
		transaction.GET("/total", transactionController.GetTotalTransactionByInterval)
		transaction.GET("/stats", transactionController.GetTransactionStats)
	}
}
