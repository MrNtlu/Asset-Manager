package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func assetRouter(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware) {
	asset := router.Group("/asset")
	{

		asset.GET("", assetController.GetAssetsByUserID)
		asset.POST("", assetController.CreateAsset)

		asset.Use(jwtToken.MiddlewareFunc())
		{
			//asset.POST("", assetController.CreateAsset)
		}
	}
}
