package routes

import (
	"asset_backend/controllers"
	"asset_backend/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func bankAccountRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	bankAccountController := controllers.NewBankAccountController(mongoDB)

	bankAccount := router.Group("/ba").Use(jwtToken.MiddlewareFunc())
	{
		bankAccount.DELETE("/all", bankAccountController.DeleteAllBankAccountsByUserID)
		bankAccount.DELETE("", bankAccountController.DeleteBankAccountByBAID)
		bankAccount.POST("", bankAccountController.CreateBankAccount)
		bankAccount.PUT("", bankAccountController.UpdateBankAccount)
		bankAccount.GET("", bankAccountController.GetBankAccountsByUserID)
		bankAccount.GET("/stats", bankAccountController.GetBankAccountStatistics)
	}
}
