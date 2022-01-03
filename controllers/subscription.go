package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubscriptionController struct{}

func (s *SubscriptionController) CreateSubscription(c *gin.Context) {
	var data requests.Subscription
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.CreateSubscription("1", data); err != nil {
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

	//uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.CreateCard("1", data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}

func (s *SubscriptionController) GetCardsByUserID(c *gin.Context) {
	//uid := jwt.ExtractClaims(c)["id"].(string)
	cards, err := models.GetCardsByUserID("1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cards})
}

func (s *SubscriptionController) GetSubscriptionsByCardID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptions, err := models.GetSubscriptionsByCardID("1", data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subscriptions})
}

func (s *SubscriptionController) GetSubscriptionsByUserID(c *gin.Context) {
	var data requests.SubscriptionSort
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptions, err := models.GetSubscriptionsByUserID("1", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subscriptions})
}

func (s *SubscriptionController) GetSubscriptionDetails(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	subscription, err := models.GetSubscriptionDetails("1", data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subscription})
}

func (s *SubscriptionController) GetSubscriptionStatisticsByUserID(c *gin.Context) {
	//uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionStats, err := models.GetSubscriptionStatisticsByUserID("1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subscriptionStats})
}

func (s *SubscriptionController) GetCardStatisticsByUserID(c *gin.Context) {
	//uid := jwt.ExtractClaims(c)["id"].(string)
	cardStats, err := models.GetCardStatisticsByUserID("1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cardStats})
}

func (s *SubscriptionController) UpdateSubscription(c *gin.Context) {
	var data requests.SubscriptionUpdate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	subscription, err := models.GetSubscriptionByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())

		return
	}

	if subscription.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "subscription not found"})
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	if "1" != subscription.UserID {
		c.JSON(http.StatusForbidden, gin.H{"message": "unauthorized access"})
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
		c.JSON(http.StatusInternalServerError, err.Error())

		return
	}

	if card.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"message": "card not found"})
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	if "1" != card.UserID {
		c.JSON(http.StatusForbidden, gin.H{"message": "unauthorized access"})
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

	if err := models.DeleteSubscriptionBySubscriptionID(data.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription deleted successfully"})
}

func (s *SubscriptionController) DeleteAllSubscriptionsByUserID(c *gin.Context) {
	//uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAllSubscriptionsByUserID("1"); err != nil {
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

	if err := models.DeleteCardByCardID(data.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "card deleted successfully"})
}

func (s *SubscriptionController) DeleteAllCardsByUserID(c *gin.Context) {
	//uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAllCardsByUserID("1"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cards deleted successfully by user id"})
}
