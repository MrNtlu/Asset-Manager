package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type LogController struct{}

// Create Log
// @Summary Create Log
// @Description Creates Log
// @Tags logs
// @Accept application/json
// @Produce application/json
// @Param log body requests.CreateLog true "Log Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {string} string
// @Router /log [post]
func (l *LogController) CreateLog(c *gin.Context) {
	var data requests.CreateLog
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	go models.CreateLog(uid, data)

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
}
