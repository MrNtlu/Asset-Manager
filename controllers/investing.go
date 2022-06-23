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

type InvestingController struct {
	Database *db.MongoDB
}

func NewInvestingController(mongoDB *db.MongoDB) InvestingController {
	return InvestingController{
		Database: mongoDB,
	}
}

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
			"error": validatorErrorHandler(err),
		})

		return
	}

	var (
		investings []responses.InvestingResponse
		cacheKey   = "investings/" + data.Type
	)

	if data.Type == "exchange" {
		result, err := db.RedisDB.Get(context.TODO(), cacheKey).Result()
		if err == nil && result != "" {
			if err := msgpack.Unmarshal([]byte(result), &investings); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": investings})

			return
		}
	}

	investingModel := models.NewInvestingModel(i.Database)

	investings, err := investingModel.GetInvestingsByTypeAndMarket(data.Type, data.Market)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if data.Type == "exchange" {
		if len(investings) > 0 {
			marshalInvestings, _ := msgpack.Marshal(investings)
			go db.RedisDB.Set(context.TODO(), cacheKey, marshalInvestings, db.RedisXLExpire)
		}
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
			"error": validatorErrorHandler(err),
		})

		return
	}

	var (
		investings []responses.InvestingTableResponse
		cacheKey   = "investings/prices/" + data.Type + "/" + data.Market
	)

	result, err := db.RedisDB.Get(context.TODO(), cacheKey).Result()
	if err != nil || result == "" {
		investingModel := models.NewInvestingModel(i.Database)

		investings, err = investingModel.GetInvestingPriceTableByTypeAndMarket(data.Type, data.Market)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if len(investings) > 0 {
			marshalInvestings, _ := msgpack.Marshal(investings)
			go db.RedisDB.Set(context.TODO(), cacheKey, marshalInvestings, db.RedisXSExpire)
		}
	} else {
		if err := msgpack.Unmarshal([]byte(result), &investings); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": investings})
}
