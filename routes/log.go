package routes

import (
	"asset_backend/controllers"
	"asset_backend/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func logRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	logController := controllers.NewLogController(mongoDB)

	log := router.Group("/log")
	{
		log.Use(jwtToken.MiddlewareFunc())
		{
			log.POST("", logController.CreateLog)
		}
	}
}
