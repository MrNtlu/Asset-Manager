package controllers

import (
	"asset_backend/db"
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vmihailenco/msgpack/v5"
)

type InvestingController struct{}

// Investings
// @Summary Get Investings by Type and Market
// @Description Returns investing list by type and market
// @Tags investing
// @Accept application/json
// @Produce application/json
// @Param investings query requests.Investings true "Investings"
// @Success 200 {array} responses.InvestingResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /investings [get]
func (i *InvestingController) GetInvestingsByTypeAndMarket(c *gin.Context) {
	var data requests.Investings
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	investings, err := models.GetInvestingsByTypeAndMarket(data.Type, data.Market)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": investings})
}

// Investing Price Table
// @Summary Get Investing Price Table by Type and Market
// @Description Returns investing price table by type and market
// @Tags investing
// @Accept application/json
// @Produce application/json
// @Param investings query requests.Investings true "Investings"
// @Success 200 {array} responses.InvestingTableResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /investings/prices [get]
func (i *InvestingController) GetInvestingPriceTableByTypeAndMarket(c *gin.Context) {
	var data requests.Investings
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	var investings []responses.InvestingTableResponse
	var cacheKey = "investings/" + data.Type + "/" + data.Market

	result, err := db.RedisDB.Get(context.TODO(), cacheKey).Result()
	if err != nil || result == "" {
		investings, err = models.GetInvestingPriceTableByTypeAndMarket(data.Type, data.Market)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if len(investings) > 0 {
			marshalInvestings, _ := msgpack.Marshal(investings)
			go db.RedisDB.Set(context.TODO(), cacheKey, marshalInvestings, db.RedisSExpire)
		}
	} else {
		msgpack.Unmarshal([]byte(result), &investings)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": investings})
}
