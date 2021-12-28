package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AssetController struct{}

func (a *AssetController) CreateAsset(c *gin.Context) {
	var data requests.Asset
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.CreateAsset(data, "1"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}

func (a *AssetController) GetAssetsByUserID(c *gin.Context) {
	//uid := jwt.ExtractClaims(c)["id"].(string)
	assets, err := models.GetAssetsByUserID("1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully fetched.", "data": assets})
}
