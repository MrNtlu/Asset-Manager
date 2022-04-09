package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/responses"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type SubscriptionController struct{}

var (
	errSubscriptionNotFound = "subscription not found"
	errSubscriptionPremium  = "free members can add up to 5 subscriptions, you can get premium membership for unlimited access"
)

// Create Subscription
// @Summary Create Subscription
// @Description Creates subscription
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param subscription body requests.Subscription true "Subscription Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription [post]
func (s *SubscriptionController) CreateSubscription(c *gin.Context) {
	var data requests.Subscription
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	isPremium := models.IsUserPremium(uid)

	if !isPremium && models.GetUserSubscriptionCount(uid) >= 5 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errSubscriptionPremium,
		})
		return
	}

	if err := models.CreateSubscription(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}

// Create Card
// @Summary Create Card
// @Description Creates card
// @Tags card
// @Accept application/json
// @Produce application/json
// @Param card body requests.Card true "Card Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /card [post]
func (s *SubscriptionController) CreateCard(c *gin.Context) {
	var data requests.Card
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.CreateCard(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
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
func (s *SubscriptionController) GetCardsByUserID(c *gin.Context) {
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

// Subscriptions By Card
// @Summary Get Subscriptions by Card ID
// @Description Returns subscriptions by card id
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param ID query requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} models.Subscription
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription/card [get]
func (s *SubscriptionController) GetSubscriptionsByCardID(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptions, err := models.GetSubscriptionsByCardID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": subscriptions})
}

// Subscriptions and Stats
// @Summary Subscriptions and Stats by User ID
// @Description Returns subscriptions and stats by user id
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param subscriptionsort query requests.SubscriptionSort true "Subscription Sort"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.SubscriptionAndStats
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription [get]
func (s *SubscriptionController) GetSubscriptionsAndStatsByUserID(c *gin.Context) {
	var data requests.SubscriptionSort
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptions, err := models.GetSubscriptionsByUserID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	subscriptionStats, err := models.GetSubscriptionStatisticsByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var subscriptionAndStats = responses.SubscriptionAndStats{
		Data:  subscriptions,
		Stats: subscriptionStats,
	}

	c.JSON(http.StatusOK, subscriptionAndStats)
}

// Subscription Details
// @Summary Get Subscription Details
// @Description Returns subscription details
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param ID query requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.SubscriptionDetails
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /subscription/details [get]
func (s *SubscriptionController) GetSubscriptionDetails(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	subscription, err := models.GetSubscriptionDetails(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if subscription.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": errSubscriptionNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": subscription})
}

// Subscription Statistics
// @Summary Get Subscription Statistics by User ID
// @Description Returns subscription statistics
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.SubscriptionStatistics
// @Failure 500 {string} string
// @Router /subscription/stats [get]
func (s *SubscriptionController) GetSubscriptionStatisticsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionStats, err := models.GetSubscriptionStatisticsByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": subscriptionStats})
}

// Card Statistics
// @Summary Get Card Statistics by User ID
// @Description Returns card statistics
// @Tags card
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.CardStatistics
// @Failure 500 {string} string
// @Router /card/stats [get]
func (s *SubscriptionController) GetCardStatisticsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	cardStats, err := models.GetCardStatisticsByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": cardStats})
}

// Update Subscription
// @Summary Update Subscription
// @Description Updates subscription
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param subscriptionupdate body requests.SubscriptionUpdate true "Subscription Update"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Couldn't find user"
// @Failure 500 {string} string
// @Router /subscription [put]
func (s *SubscriptionController) UpdateSubscription(c *gin.Context) {
	var data requests.SubscriptionUpdate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	subscription, err := models.GetSubscriptionByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if subscription.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": errSubscriptionNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != subscription.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	if err := models.UpdateSubscription(data, subscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription updated"})
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
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Couldn't find user"
// @Failure 500 {string} string
// @Router /card [put]
func (s *SubscriptionController) UpdateCard(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != card.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	if err := models.UpdateCard(data, card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "card updated"})
}

// Delete Subscription By ID
// @Summary Delete subscription by subscription id
// @Description Deletes subscription by id
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param ID body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /subscription [delete]
func (s *SubscriptionController) DeleteSubscriptionBySubscriptionID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	isDeleted, err := models.DeleteSubscriptionBySubscriptionID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if isDeleted {
		c.JSON(http.StatusOK, gin.H{"message": "subscription deleted successfully"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": ErrUnauthorized})
}

// Delete All Subscriptions
// @Summary Delete all subscriptions by user id
// @Description Deletes all subscriptions by user id
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /subscription/all [delete]
func (s *SubscriptionController) DeleteAllSubscriptionsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAllSubscriptionsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscriptions deleted successfully by user id"})
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
func (s *SubscriptionController) DeleteCardByCardID(c *gin.Context) {
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
		go models.UpdateSubscriptionCardIDToNull(uid, &data.ID)
		c.JSON(http.StatusOK, gin.H{"message": "card deleted successfully"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"error": "unauthorized delete"})
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
func (s *SubscriptionController) DeleteAllCardsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAllCardsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go models.UpdateSubscriptionCardIDToNull(uid, nil)
	c.JSON(http.StatusOK, gin.H{"message": "cards deleted successfully by user id"})
}
