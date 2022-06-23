package routes

import (
	"asset_backend/controllers"
	"asset_backend/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func cardRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	cardController := controllers.NewCardController(mongoDB)

	card := router.Group("/card").Use(jwtToken.MiddlewareFunc())
	{
		card.DELETE("/all", cardController.DeleteAllCardsByUserID)
		card.DELETE("", cardController.DeleteCardByCardID)
		card.PUT("", cardController.UpdateCard)
		card.POST("", cardController.CreateCard)
		card.GET("", cardController.GetCardsByUserID)
		card.GET("/stats", cardController.GetCardStatisticsByUserIDAndCardID)
	}
}
