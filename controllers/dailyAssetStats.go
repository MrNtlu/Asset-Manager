package controllers

import (
	"asset_backend/db"
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/vmihailenco/msgpack/v5"
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
			"error": validatorErrorHandler(err),
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

	var (
		cacheKey        = "daily-asset/" + uid + "/" + data.Interval
		dailyAssetStats responses.DailyAssetStats
	)

	result, err := db.RedisDB.Get(context.TODO(), cacheKey).Result()
	if err != nil || result == "" {
		dailyAssetStats, err = models.GetAssetStatsByUserID(uid, data.Interval)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		marshalDailyAssetStats, _ := msgpack.Marshal(dailyAssetStats)
		go db.RedisDB.Set(context.TODO(), cacheKey, marshalDailyAssetStats, db.RedisLExpire)
	} else {
		if err := msgpack.Unmarshal([]byte(result), &dailyAssetStats); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": dailyAssetStats})
}
