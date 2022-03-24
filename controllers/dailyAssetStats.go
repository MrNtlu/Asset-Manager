package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type DailyAssetStatsController struct{}

// Daily Asset Stats
// @Summary Get Daily Asset Stats by User ID
// @Description Returns daily asset stats by user id
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Param dailyassetstatsinterval query requests.DailyAssetStatsInterval true "Daily Asset Stats Interval"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.DailyAssetStats
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asset/daily-stats [get]
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
