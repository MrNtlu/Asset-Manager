package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func bankAccountRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware) {
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
