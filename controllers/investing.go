package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InvestingController struct{}

func (i *InvestingController) GetInvestingsByType(c *gin.Context) {
	var data requests.Type
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	investings, err := models.GetInvestingsByType(data.AssetType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": investings.Data, "pagination": investings.Pagination})
	//c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": investings})
}
