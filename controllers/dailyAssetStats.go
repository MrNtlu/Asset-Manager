package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type DailyAssetStatsController struct{}

func (d *DailyAssetStatsController) GetAssetStatsByUserID(c *gin.Context) {
	var data requests.DailyAssetStatsInterval
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	isPremium := models.IsUserPremium(uid)

	if !isPremium && data.Interval != "weekly" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errPremiumFeature,
		})

		return
	}

	dailyAssetStats, err := models.GetAssetStatsByUserID(uid, data.Interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dailyAssetStats})
}
