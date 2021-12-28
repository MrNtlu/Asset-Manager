package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func userRouter(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", jwtToken.LoginHandler)
	}

	user := router.Group("/user")
	{
		// user.GET("", controller.)

		user.Use(jwtToken.MiddlewareFunc()) //Auth required
		{

		}
	}
}
