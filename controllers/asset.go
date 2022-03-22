package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type AssetController struct{}

var (
	errAssetNotFound = "asset not found"
	errAssetPremium  = "free members can add up to 10, you can get premium membership for unlimited access"
)

func (a *AssetController) CreateAsset(c *gin.Context) {
	var data requests.AssetCreate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	isPremium := models.IsUserPremium(uid)
	assetCount := models.GetUserAssetCount(uid)

	if !isPremium && assetCount >= 10 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errAssetPremium,
		})
		return
	}

	if err := models.CreateAsset(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}

func (a *AssetController) GetAssetsAndStatsByUserID(c *gin.Context) {
	var data requests.AssetSort
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	assets, err := models.GetAssetsByUserID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	assetStat, err := models.GetAllAssetStats(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": assets, "stats": assetStat})
}

func (a *AssetController) GetAssetStatsByAssetAndUserID(c *gin.Context) {
	var data requests.AssetDetails
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}
	uid := jwt.ExtractClaims(c)["id"].(string)

	assetDetails, err := models.GetAssetStatsByAssetAndUserID(uid, data.ToAsset, data.FromAsset, data.AssetMarket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": assetDetails})
}

func (a *AssetController) GetAllAssetStatsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	assetStat, err := models.GetAllAssetStats(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": assetStat})
}

func (a *AssetController) GetAssetLogsByUserID(c *gin.Context) {
	var data requests.AssetLog
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	assets, pagination, err := models.GetAssetLogsByUserID(uid, data)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if asset.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": errAssetNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != asset.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
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

	uid := jwt.ExtractClaims(c)["id"].(string)
	isDeleted, err := models.DeleteAssetLogByAssetID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if isDeleted {
		c.JSON(http.StatusOK, gin.H{"message": "asset deleted successfully"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": ErrUnauthorized})
}

func (a *AssetController) DeleteAssetLogsByUserID(c *gin.Context) {
	var data requests.AssetLogsDelete
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAssetLogsByUserID(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "assets deleted successfully"})
}

func (a *AssetController) DeleteAllAssetsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAllAssetsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "assets deleted successfully by user id"})
}
