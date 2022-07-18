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

type FavouriteInvestingController struct {
	Database *db.MongoDB
}

func NewFavouriteInvestingController(mongoDB *db.MongoDB) FavouriteInvestingController {
	return FavouriteInvestingController{
		Database: mongoDB,
	}
}

var (
	errFavInvestingPremium = "Free members can add up to 5, you can get premium membership to increase the limit."
	errFavInvestingLimit   = "You've reached the limit."
)

// Create Favourite Investing
// @Summary Create Favourite Investing
// @Description Creates favourite investing
// @Tags favouriteinvesting
// @Accept application/json
// @Produce application/json
// @Param favouriteinvestingcreate body requests.FavouriteInvestingCreate true "Favourite Investing Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {string} string
// @Failure 403 {string} string
// @Failure 500 {string} string
// @Router /watchlist [post]
func (fi *FavouriteInvestingController) CreateFavouriteInvesting(c *gin.Context) {
	var data requests.FavouriteInvestingCreate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userModel := models.NewUserModel(fi.Database)
	isPremium := userModel.IsUserPremium(uid)

	favouriteInvestingModel := models.NewFavouriteInvestingModel(fi.Database)

	count := favouriteInvestingModel.GetFavouriteInvestingsCount(uid)
	if !isPremium && count >= 5 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errFavInvestingPremium,
		})

		return
	} else if isPremium && count >= 10 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errFavInvestingLimit,
		})

		return
	}

	if err := favouriteInvestingModel.CreateFavouriteInvesting(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	go db.RedisDB.Del(context.TODO(), ("watchlist/" + uid))

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}

// Favourite Investings By User ID
// @Summary Get Favourite Investings by User ID
// @Description Returns favourite investings by user id
// @Tags favouriteinvesting
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.FavouriteInvesting
// @Failure 500 {string} string
// @Router /watchlist [get]
func (fi *FavouriteInvestingController) GetFavouriteInvestings(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	var (
		cacheKey            = "watchlist/" + uid
		favouriteInvestings []responses.FavouriteInvesting
	)

	result, err := db.RedisDB.Get(context.TODO(), cacheKey).Result()
	if err != nil || result == "" {
		favouriteInvestingModel := models.NewFavouriteInvestingModel(fi.Database)
		favouriteInvestings, err = favouriteInvestingModel.GetFavouriteInvestings(uid)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		marshalInvestings, _ := msgpack.Marshal(favouriteInvestings)
		go db.RedisDB.Set(context.TODO(), cacheKey, marshalInvestings, db.RedisXLExpire)
	} else if err := msgpack.Unmarshal([]byte(result), &favouriteInvestings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": favouriteInvestings})
}

// Delete Favourite Investing By ID
// @Summary Delete favourite investing by favourite investing id
// @Description Deletes favourite investing by id
// @Tags favouriteinvesting
// @Accept application/json
// @Produce application/json
// @Param ID body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /watchlist [delete]
func (fi *FavouriteInvestingController) DeleteFavouriteInvestingByID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	favouriteInvestingModel := models.NewFavouriteInvestingModel(fi.Database)

	isDeleted, err := favouriteInvestingModel.DeleteFavouriteInvestingByID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		go db.RedisDB.Del(context.TODO(), ("watchlist/" + uid))

		c.JSON(http.StatusOK, gin.H{"message": "Watchlist deleted successfully."})

		return
	}

	c.JSON(http.StatusOK, gin.H{"error": "Unauthorized delete."})
}

// Delete All Favourite Investing
// @Summary Delete all favourite investings by user id
// @Description Deletes favourite investing by user id
// @Tags favouriteinvesting
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /watchlist/all [delete]
func (fi *FavouriteInvestingController) DeleteAllFavouriteInvestingsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	favouriteInvestingModel := models.NewFavouriteInvestingModel(fi.Database)
	if err := favouriteInvestingModel.DeleteAllFavouriteInvestingsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	go db.RedisDB.Del(context.TODO(), ("watchlist/" + uid))
	c.JSON(http.StatusOK, gin.H{"message": "Watchlist deleted successfully."})
}
