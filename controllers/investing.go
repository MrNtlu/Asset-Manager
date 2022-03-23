package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InvestingController struct{}

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
