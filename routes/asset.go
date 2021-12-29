package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func assetRouter(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware) {
	asset := router.Group("/asset")
	{

		asset.DELETE("/log", assetController.DeleteAssetLogByAssetID)
		asset.DELETE("/logs", assetController.DeleteAssetLogsByUserID)
		asset.DELETE("", assetController.DeleteAllAssetsByUserID)
		asset.GET("/logs", assetController.GetAssetLogsByUserID)
		asset.GET("", assetController.GetAssetsByUserID)
		asset.PUT("", assetController.UpdateAssetLogByAssetID)
		asset.POST("", assetController.CreateAsset)

		asset.Use(jwtToken.MiddlewareFunc())
		{
			//asset.POST("", assetController.CreateAsset)
		}
	}
}
