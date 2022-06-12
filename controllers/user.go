package controllers

import (
	"asset_backend/db"
	"asset_backend/helpers"
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/responses"
	"asset_backend/utils"
	"context"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sethvargo/go-password/password"
)

type UserController struct{}

var (
	errAlreadyRegistered = "User already registered."
	errPasswordNoMatch   = "Passwords do not match."
	errNoUser            = "Sorry, couldn't find user."
	errOAuthUser         = "Sorry, you can't do this action."
	errMailAlreadySent   = "Password reset mail already sent, you have to wait 5 minutes before sending another. Please check spam mails."
	errPremiumFeature    = "This feature requires premium membership."
)

// Register
// @Summary User Registration
// @Description Allows users to register
// @Tags auth
// @Accept application/json
// @Produce application/json
// @Param register body requests.Register true "User registration info"
// @Success 201 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /auth/register [post]
func (u *UserController) Register(c *gin.Context) {
	var data requests.Register
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	user, _ := models.FindUserByEmail(data.EmailAddress)

	if user.EmailAddress != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errAlreadyRegistered,
		})

		return
	}

	if err := models.CreateUser(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registered successfully."})
}

// Change Currency
// @Summary Change User Currency
// @Description Users can change their default currency
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param changecurrency body requests.ChangeCurrency true "Set currency"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/change-currency [put]
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

	go db.RedisDB.Del(context.TODO(), ("asset/" + uid))

	c.JSON(http.StatusOK, gin.H{"message": "Successfully changed currency."})
}

// Update FCM Token
// @Summary Updates FCM User Token
// @Description Depending on logged in device fcm token will be updated
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param changefcmtoken body requests.ChangeFCMToken true "Set token"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/update-token [put]
func (u *UserController) UpdateFCMToken(c *gin.Context) {
	var data requests.ChangeFCMToken
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

	if user.FCMToken != data.FCMToken {
		user.FCMToken = data.FCMToken
		if err = models.UpdateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated FCM Token."})
}

// Change User Membership
// @Summary Change User Membership
// @Description User membership status will be updated depending on subscription status
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param changemembership body requests.ChangeMembership true "Set Membership"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /user/membership [put]
func (u *UserController) ChangeUserMembership(c *gin.Context) {
	var data requests.ChangeMembership
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	if err := models.UpdateUserMembership(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated membership."})
}

// Change Password
// @Summary Change User Password
// @Description Users can change their password
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param ChangePassword body requests.ChangePassword true "Set new password"
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /user/change-password [put]
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

	if user.IsOAuthUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errOAuthUser,
		})
		return
	}

	if err = utils.CheckPassword([]byte(user.Password), []byte(data.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"error": errPasswordNoMatch},
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

	c.JSON(http.StatusOK, gin.H{"message": "Successfully changed password."})
}

// Forgot Password
// @Summary Will be used when user forgot password
// @Description Users can change their password when they forgot
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param ForgotPassword body requests.ForgotPassword true "User's email"
// @Success 200 {string} string
// @Failure 400 {string} string "Couldn't find any user"
// @Failure 500 {string} string
// @Router /user/forgot-password [post]
func (u *UserController) ForgotPassword(c *gin.Context) {
	var data requests.ForgotPassword
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	user, err := models.FindUserByEmail(data.EmailAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errNoUser,
		})
		return
	}

	if user.EmailAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errNoUser,
		})

		return
	}

	if user.IsOAuthUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errOAuthUser,
		})
		return
	}

	var resetToken string
	if user.PasswordResetToken == "" {
		resetToken = uuid.NewString()
		user.PasswordResetToken = resetToken
		if err = models.UpdateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		time.AfterFunc(5*time.Minute, func() {
			user.PasswordResetToken = ""
			go models.UpdateUser(user)
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errMailAlreadySent,
		})

		return
	}

	if err := helpers.SendForgotPasswordEmail(resetToken, user.EmailAddress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully send password reset email."})
}

func (u *UserController) ConfirmPasswordReset(c *gin.Context) {
	token := c.Query("token")
	email := c.Query("mail")

	user, err := models.FindUserByResetTokenAndEmail(token, email)
	if err != nil {
		http.ServeFile(c.Writer, c.Request, "assets/error_password_reset.html")
		return
	}

	if user.EmailAddress == "" {
		http.ServeFile(c.Writer, c.Request, "assets/error_password_reset.html")
		return
	}

	if user.IsOAuthUser {
		http.ServeFile(c.Writer, c.Request, "assets/error_password_reset.html")
		return
	}

	generatedPass, err := password.Generate(10, 4, 0, true, false)
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

	if err := helpers.SendPasswordChangedEmail(generatedPass, user.EmailAddress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	http.ServeFile(c.Writer, c.Request, "assets/confirm_password.html")
}

// User Info
// @Summary User membership info
// @Description Returns users membership & investing/subscription limits
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.UserInfo "User Info"
// @Router /user/info [get]
func (u *UserController) GetUserInfo(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	info, _ := models.FindUserByID(uid)
	assetCount := models.GetUserAssetCount(uid)
	subscritionCount := models.GetUserSubscriptionCount(uid)

	var investingLimit string
	var subscriptionLimit string
	if info.IsPremium {
		investingLimit = fmt.Sprintf("%v", assetCount) + "/∞"
		subscriptionLimit = fmt.Sprintf("%v", subscritionCount) + "/∞"
	} else {
		investingLimit = fmt.Sprintf("%v", assetCount) + "/10"
		subscriptionLimit = fmt.Sprintf("%v", subscritionCount) + "/5"
	}

	userInfo := responses.UserInfo{
		IsPremium:         info.IsPremium,
		IsLifetimePremium: info.IsLifetimePremium,
		IsOAuth:           info.IsOAuthUser,
		EmailAddress:      info.EmailAddress,
		Currency:          info.Currency,
		FCMToken:          info.FCMToken,
		InvestingLimit:    investingLimit,
		SubscriptionLimit: subscriptionLimit,
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched user info.", "data": userInfo})
}

// Delete User
// @Summary Deletes user information
// @Description Deletes everything related to user
// @Tags user
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Error 500 {string} string
// @Router /user [delete]
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
	go models.DeleteAllAssetStatsByUserID(uid)
	go models.DeleteAllLogsByUserID(uid)
	go models.DeleteAllTransactionsByUserID(uid)
	go models.DeleteAllBankAccountsByUserID(uid)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted user."})
}
