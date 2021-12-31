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
	if err := models.CreateSubscription(data, "1"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}

func (s *SubscriptionController) CreateCard(c *gin.Context) {
	var data requests.Card
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	//uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.CreateCard(data, "1"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}
