package controllers

import (
	"asset_backend/db"
	"asset_backend/models"
	"asset_backend/requests"
	"context"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type CardController struct{}

var (
	errCardPremium  = "Free members can add up to 3 credit cards, you can get premium membership for unlimited access."
	errNoCreditCard = "Couldn't find credit card."
)

// Create Card
// @Summary Create Card
// @Description Creates card
// @Tags card
// @Accept application/json
// @Produce application/json
// @Param card body requests.Card true "Card Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.Card
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /card [post]
func (cc *CardController) CreateCard(c *gin.Context) {
	var data requests.Card
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	isPremium := models.IsUserPremium(uid)

	if !isPremium && models.GetUserCardCount(uid) >= 3 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errCardPremium,
		})
		return
	}

	var (
		createdCard models.Card
		err         error
	)
	if createdCard, err = models.CreateCard(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go db.RedisDB.Del(context.TODO(), ("card/" + uid))

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdCard})
}

// Cards By User ID
// @Summary Get Cards by User ID
// @Description Returns cards by user id
// @Tags card
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} models.Card
// @Failure 500 {string} string
// @Router /card [get]
func (cc *CardController) GetCardsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	cards, err := models.GetCardsByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": cards})
}

// Card Statistics
// @Summary Get Card Statistics by User ID & Card ID
// @Description Returns card statistics
// @Tags card
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.CardStatistics
// @Failure 500 {string} string
// @Router /card/stats [get]
func (cc *CardController) GetCardStatisticsByUserIDAndCardID(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	cardStats, err := models.GetCardStatisticsByUserIDAndCardID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": cardStats})
}

// Update Card
// @Summary Update Card
// @Description Updates card
// @Tags card
// @Accept application/json
// @Produce application/json
// @Param cardupdate body requests.CardUpdate true "Card Update"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.Card
// @Failure 400 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Couldn't find user"
// @Failure 500 {string} string
// @Router /card [put]
func (cc *CardController) UpdateCard(c *gin.Context) {
	var data requests.CardUpdate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	card, err := models.GetCardByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if card.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": errNoCreditCard})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != card.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	var updatedCard models.Card
	if updatedCard, err = models.UpdateCard(data, card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go db.RedisDB.Del(context.TODO(), ("card/" + uid))

	c.JSON(http.StatusOK, gin.H{"message": "Card updated.", "data": updatedCard})
}

// Delete Card By ID
// @Summary Delete card by card id
// @Description Deletes card by id
// @Tags card
// @Accept application/json
// @Produce application/json
// @Param ID body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /card [delete]
func (cc *CardController) DeleteCardByCardID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	isDeleted, err := models.DeleteCardByCardID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if isDeleted {
		go db.RedisDB.Del(context.TODO(), ("card/" + uid))
		go models.UpdateSubscriptionCardIDToNull(uid, &data.ID)
		go models.UpdateTransactionMethodIDToNull(uid, &data.ID, models.CreditCard)
		c.JSON(http.StatusOK, gin.H{"message": "Card deleted successfully."})

		return
	}

	c.JSON(http.StatusOK, gin.H{"error": "Unauthorized delete."})
}

// Delete All Cards
// @Summary Delete all cards by user id
// @Description Deletes all cards by user id
// @Tags card
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /card/all [delete]
func (cc *CardController) DeleteAllCardsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAllCardsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go models.UpdateSubscriptionCardIDToNull(uid, nil)
	go models.UpdateTransactionMethodIDToNull(uid, nil, models.CreditCard)
	go db.RedisDB.Del(context.TODO(), ("card/" + uid))
	c.JSON(http.StatusOK, gin.H{"message": "Cards deleted successfully by user id."})
}
