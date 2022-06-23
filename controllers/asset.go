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
	"golang.org/x/sync/errgroup"
)

type AssetController struct {
	Database *db.MongoDB
}

func NewAssetController(mongoDB *db.MongoDB) AssetController {
	return AssetController{
		Database: mongoDB,
	}
}

var (
	errAssetNotFound = "Asset not found."
	errAssetPremium  = "Free members can add up to 10, you can get premium membership for unlimited access."
)

// Create Asset
// @Summary Create Asset
// @Description Creates asset
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Param assetcreate body requests.AssetCreate true "Asset Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {string} string
// @Failure 500 {string} string
// @Router /asset [post]
func (a *AssetController) CreateAsset(c *gin.Context) {
	var data requests.AssetCreate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userModel := models.NewUserModel(a.Database)
	isPremium := userModel.IsUserPremium(uid)

	assetModel := models.NewAssetModel(a.Database)
	if !isPremium && assetModel.GetUserAssetCount(uid) >= 10 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errAssetPremium,
		})

		return
	}

	if err := assetModel.CreateAsset(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}

// Create Asset Log
// @Summary Create Asset Log
// @Description Creates asset log
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Param assetcreate body requests.AssetCreate true "Asset Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {string} string
// @Failure 500 {string} string
// @Router /asset/log [post]
func (a *AssetController) CreateAssetLog(c *gin.Context) {
	var data requests.AssetCreate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	assetModel := models.NewAssetModel(a.Database)

	if err := assetModel.CreateAsset(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}

// Assets & Stats by User ID
// @Summary Get Assets & Stats by User ID
// @Description Returns assets and stats by user id
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Param assetsort query requests.AssetSortFilter true "Asset Sort"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.AssetAndStats
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asset [get]
func (a *AssetController) GetAssetsAndStatsByUserID(c *gin.Context) {
	var data requests.AssetSortFilter
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	assetModel := models.NewAssetModel(a.Database)

	var (
		assetAndStats responses.AssetAndStats
		assets        []responses.Asset
		assetStat     responses.AssetStats
		err           error
		errs, _       = errgroup.WithContext(context.TODO())
	)

	errs.Go(func() error {
		assets, err = assetModel.GetAssetsByUserID(uid, data)
		return err
	})

	errs.Go(func() error {
		assetStat, err = assetModel.GetAllAssetStats(uid)
		return err
	})

	err = errs.Wait()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	assetAndStats = responses.AssetAndStats{
		Data:  assets,
		Stats: assetStat,
	}

	c.JSON(http.StatusOK, assetAndStats)
}

// Asset Stats
// @Summary Get Asset Stats by Asset and User ID
// @Description Returns asset stats by asset and user id
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Param assetdetails query requests.AssetDetails true "Asset Details"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.AssetDetails
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asset/details [get]
func (a *AssetController) GetAssetStatsByAssetAndUserID(c *gin.Context) {
	var data requests.AssetDetails
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	assetModel := models.NewAssetModel(a.Database)

	assetDetails, err := assetModel.GetAssetStatsByAssetAndUserID(uid, data.ToAsset, data.FromAsset, data.AssetMarket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": assetDetails})
}

// All Asset Stats
// @Summary Get Asset Stats by User ID
// @Description Returns asset stats by user id
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.AssetStats
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asset/stats [get]
func (a *AssetController) GetAllAssetStatsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	assetModel := models.NewAssetModel(a.Database)

	assetStat, err := assetModel.GetAllAssetStats(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": assetStat})
}

// Asset Logs
// @Summary Get Asset Logs by User ID
// @Description Returns asset logs by user id
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Param assetlog query requests.AssetLog true "Asset Log"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} models.Asset
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asset/logs [get]
func (a *AssetController) GetAssetLogsByUserID(c *gin.Context) {
	var data requests.AssetLog
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	assetModel := models.NewAssetModel(a.Database)

	assets, pagination, err := assetModel.GetAssetLogsByUserID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": assets, "pagination": pagination})
}

// Update Asset Log
// @Summary Update Asset Log by AssetID
// @Description Updates asset log by id
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Param assetupdate body requests.AssetUpdate true "Asset Update"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Couldn't find user"
// @Failure 500 {string} string
// @Router /asset [put]
func (a *AssetController) UpdateAssetLogByAssetID(c *gin.Context) {
	var data requests.AssetUpdate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	assetModel := models.NewAssetModel(a.Database)

	asset, err := assetModel.GetAssetByID(data.ID)
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

	if err := assetModel.UpdateAssetLogByAssetID(data, asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Asset updated."})
}

// Delete Asset Log
// @Summary Delete Asset Log by AssetID
// @Description Deletes asset log by id
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Param ID body requests.ID true "Asset ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asset/log [delete]
func (a *AssetController) DeleteAssetLogByAssetID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	assetModel := models.NewAssetModel(a.Database)

	isDeleted, err := assetModel.DeleteAssetLogByAssetID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		c.JSON(http.StatusOK, gin.H{"message": "Asset deleted successfully."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": ErrUnauthorized})
}

// Delete Asset Logs
// @Summary Delete Asset Logs by User ID
// @Description Deletes all asset logs by user id
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Param assetlogsdelete body requests.AssetLogsDelete true "Asset Logs Delete"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /asset/logs [delete]
func (a *AssetController) DeleteAssetLogsByUserID(c *gin.Context) {
	var data requests.AssetLogsDelete
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	assetModel := models.NewAssetModel(a.Database)

	if err := assetModel.DeleteAssetLogsByUserID(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assets deleted successfully."})
}

// Delete All Assets
// @Summary Delete All Assets by User ID
// @Description Deletes all assets by user id
// @Tags asset
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /asset [delete]
func (a *AssetController) DeleteAllAssetsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	assetModel := models.NewAssetModel(a.Database)

	if err := assetModel.DeleteAllAssetsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assets deleted successfully by user id."})
}
