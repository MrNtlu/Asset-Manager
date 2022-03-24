package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	"github.com/gin-gonic/gin"
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

	investings, err := models.GetInvestingPriceTableByTypeAndMarket(data.Type, data.Market)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": investings})
}
