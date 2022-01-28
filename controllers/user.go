package controllers

import (
	"asset_backend/helpers"
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/utils"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sethvargo/go-password/password"
)

type UserController struct{}

func (u *UserController) Register(c *gin.Context) {
	var data requests.Register
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	user, _ := models.FindUserByEmail(data.EmailAddress)

	if user.EmailAddress != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user already registered",
		})

		return
	}

	if err := models.CreateUser(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "registered successfully"})
}

func (u *UserController) ChangeCurrency(c *gin.Context) {
	var data requests.ChangeCurrency
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	user, err := models.FindUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	user.Currency = data.Currency
	if err = models.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully changed currency"})
}

func (u *UserController) ChangePassword(c *gin.Context) {
	var data requests.ChangePassword
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	user, err := models.FindUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err = utils.CheckPassword([]byte(user.Password), []byte(data.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"error": "passwords do not match"},
		})
		return
	}

	user.Password = utils.HashPassword(data.NewPassword)
	if err = models.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully changed password"})
}

func (u *UserController) ForgotPassword(c *gin.Context) {
	var data requests.ForgotPassword
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	user, err := models.FindUserByEmail(data.EmailAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if user.EmailAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "couldn't find user",
		})

		return
	}

	resetToken := uuid.NewString()
	user.PasswordResetToken = resetToken
	if err = models.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	helpers.SendForgotPasswordEmail(resetToken, user.EmailAddress)

	c.JSON(http.StatusOK, gin.H{"message": "successfully send password reset email"})
}

func (u *UserController) ConfirmPasswordReset(c *gin.Context) {
	token := c.Query("token")
	email := c.Query("mail")

	user, err := models.FindUserByResetTokenAndEmail(token, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if user.EmailAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "couldn't find user",
		})

		return
	}

	generatedPass, err := password.Generate(10, 4, 1, true, false)
	if err != nil {
		generatedPass = user.EmailAddress + "_Password"
	}

	user.Password = utils.HashPassword(generatedPass)
	user.PasswordResetToken = ""
	if err = models.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	helpers.SendPasswordChangedEmail(generatedPass, user.EmailAddress)

	c.JSON(http.StatusOK, gin.H{"message": "new password sent via email"})
}

func (u *UserController) DeleteUser(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	if err := models.DeleteUserByID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go models.DeleteAllAssetsByUserID(uid)
	go models.DeleteAllCardsByUserID(uid)
	go models.DeleteAllSubscriptionsByUserID(uid)

	c.JSON(http.StatusOK, gin.H{"message": "successfully deleted user"})
}
