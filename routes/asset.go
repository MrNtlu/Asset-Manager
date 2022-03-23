package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func assetRouter(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware) {
	asset := router.Group("/asset").Use(jwtToken.MiddlewareFunc())
	{
		asset.DELETE("/log", assetController.DeleteAssetLogByAssetID)
		asset.DELETE("/logs", assetController.DeleteAssetLogsByUserID)
		asset.DELETE("", assetController.DeleteAllAssetsByUserID)
		asset.PUT("", assetController.UpdateAssetLogByAssetID)
		asset.POST("", assetController.CreateAsset)
		asset.GET("/details", assetController.GetAssetStatsByAssetAndUserID)
		asset.GET("/daily-stats", dailyAssetStatsController.GetAssetStatsByUserID)
		asset.GET("/stats", assetController.GetAllAssetStatsByUserID)
		asset.GET("/logs", assetController.GetAssetLogsByUserID)
		asset.GET("", assetController.GetAssetsAndStatsByUserID)
	}

	investing := router.Group("/investings")
	{
		investing.GET("", investingController.GetInvestingsByTypeAndMarket)
		investing.GET("/prices", investingController.GetInvestingPriceTableByTypeAndMarket)
	}
}
