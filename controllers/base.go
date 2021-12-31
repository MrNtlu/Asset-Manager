package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func bindJSONData(data interface{}, c *gin.Context) bool {
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return true
	}

	return false
}
