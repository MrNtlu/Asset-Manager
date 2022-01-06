package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AssetController struct{}

func (a *AssetController) CreateAsset(c *gin.Context) {
	var data requests.AssetCreate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.CreateAsset(data, "1"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}

func (a *AssetController) GetAssetsByUserID(c *gin.Context) {
	var data requests.AssetSort
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	assets, err := models.GetAssetsByUserID("1", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": assets})
}

func (a *AssetController) GetAssetLogsByUserID(c *gin.Context) {
	var data requests.AssetLog
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	assets, pagination, err := models.GetAssetLogsByUserID("1", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": assets, "pagination": pagination})
}

func (a *AssetController) UpdateAssetLogByAssetID(c *gin.Context) {
	var data requests.AssetUpdate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	asset, err := models.GetAssetByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())

		return
	}

	if asset.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "asset not found"})
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	if "1" != asset.UserID {
		c.JSON(http.StatusForbidden, gin.H{"message": "unauthorized access"})
		return
	}

	if err := models.UpdateAssetLogByAssetID(data, asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "asset updated"})
}

func (a *AssetController) DeleteAssetLogByAssetID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	if err := models.DeleteAssetLogByAssetID(data.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "asset deleted successfully"})
}

func (a *AssetController) DeleteAssetLogsByUserID(c *gin.Context) {
	var data requests.AssetLogsDelete
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAssetLogsByUserID("1", data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "assets deleted successfully"})
}

func (a *AssetController) DeleteAllAssetsByUserID(c *gin.Context) {
	//uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAllAssetsByUserID("1"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "assets deleted successfully by user id"})
}
