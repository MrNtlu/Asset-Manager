package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type SubscriptionController struct{}

var (
	errSubscriptionNotFound = "subscription not found"
	errSubscriptionPremium  = "free members can add up to 5 subscriptions, you can get premium membership for unlimited access"
)

func (s *SubscriptionController) CreateSubscription(c *gin.Context) {
	var data requests.Subscription
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	isPremium := models.IsUserPremium(uid)

	if !isPremium && models.GetUserSubscriptionCount(uid) > 5 {
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

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": subscriptions, "stats": subscriptionStats})
}

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
