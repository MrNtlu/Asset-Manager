package routes

import (
	"asset_backend/controllers"
	"asset_backend/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func favouriteInvestingRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	favInvestingController := controllers.NewFavouriteInvestingController(mongoDB)

	favInvesting := router.Group("/watchlist").Use(jwtToken.MiddlewareFunc())
	{
		favInvesting.DELETE("/all", favInvestingController.DeleteAllFavouriteInvestingsByUserID)
		favInvesting.DELETE("", favInvestingController.DeleteFavouriteInvestingByID)
		favInvesting.POST("", favInvestingController.CreateFavouriteInvesting)
		favInvesting.PUT("", favInvestingController.UpdateFavouriteInvestingOrder)
		favInvesting.GET("", favInvestingController.GetFavouriteInvestings)
	}
}
