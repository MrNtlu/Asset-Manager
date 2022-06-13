package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func oauth2Router(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware) {
	google := router.Group("")
	{
		google.GET("/login", OAuth2Controller.GoogleLogin)
		google.GET("/callback", OAuth2Controller.GoogleCallback(jwtToken))
	}

	oauth := router.Group("/oauth")
	{
		oauth.POST("/google", OAuth2Controller.OAuth2GoogleLogin(jwtToken))
		oauth.POST("/apple", OAuth2Controller.OAuth2AppleLogin(jwtToken))
	}
}
