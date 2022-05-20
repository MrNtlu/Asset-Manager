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

type BankAccountController struct{}

var (
	errBankAccountPremium = "Free members can add up to 2 bank accounts, you can get premium membership for unlimited access."
	errNoBankAccount      = "Couldn't find bank account."
)

// Create BankAccount
// @Summary Create BankAccount
// @Description Creates bank account
// @Tags bankaccount
// @Accept application/json
// @Produce application/json
// @Param bankaccount body requests.BankAccountCreate true "Bank Account Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.BankAccount
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /ba [post]
func (ba *BankAccountController) CreateBankAccount(c *gin.Context) {
	var data requests.BankAccountCreate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	isPremium := models.IsUserPremium(uid)

	if !isPremium && models.GetUserBankAccountCount(uid) >= 2 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errBankAccountPremium,
		})
		return
	}

	var (
		createdBankAccount models.BankAccount
		err                error
	)
	if createdBankAccount, err = models.CreateBankAccount(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go db.RedisDB.Del(context.TODO(), ("ba/" + uid))

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdBankAccount})
}

// BankAccounts By User ID
// @Summary Get Bank Accounts by User ID
// @Description Returns bank accounts by user id
// @Tags bankaccount
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} models.BankAccount
// @Failure 500 {string} string
// @Router /ba [get]
func (ba *BankAccountController) GetBankAccountsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	bankAccounts, err := models.GetBankAccountsByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": bankAccounts})
}

// Update BankAccount
// @Summary Update Bank Account
// @Description Updates bank account
// @Tags bankaccount
// @Accept application/json
// @Produce application/json
// @Param bankaccountupdate body requests.BankAccountUpdate true "BankAccount Update"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.BankAccount
// @Failure 400 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Couldn't find user"
// @Failure 500 {string} string
// @Router /ba [put]
func (ba *BankAccountController) UpdateBankAccount(c *gin.Context) {
	var data requests.BankAccountUpdate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	bankAccount, err := models.GetBankAccountByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if bankAccount.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": errNoBankAccount})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != bankAccount.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	var updatedBankAccount models.BankAccount
	if updatedBankAccount, err = models.UpdateBankAccount(data, bankAccount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go db.RedisDB.Del(context.TODO(), ("ba/" + uid))

	c.JSON(http.StatusOK, gin.H{"message": "Bank account updated.", "data": updatedBankAccount})
}

// Delete BankAccount By ID
// @Summary Delete bank account by bank account id
// @Description Deletes bank account by id
// @Tags bankaccount
// @Accept application/json
// @Produce application/json
// @Param ID body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /ba [delete]
func (ba *BankAccountController) DeleteBankAccountByBAID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	isDeleted, err := models.DeleteBankAccountByBAID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if isDeleted {
		go db.RedisDB.Del(context.TODO(), ("ba/" + uid))
		go models.UpdateTransactionMethodIDToNull(uid, &data.ID, models.BankAcc)
		c.JSON(http.StatusOK, gin.H{"message": "Bank account deleted successfully."})

		return
	}

	c.JSON(http.StatusOK, gin.H{"error": "Unauthorized delete."})
}

// Delete All BankAccounts
// @Summary Delete all bank accounts by user id
// @Description Deletes all bank accounts by user id
// @Tags bankaccount
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /ba/all [delete]
func (ba *BankAccountController) DeleteAllBankAccountsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	if err := models.DeleteAllBankAccountsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go models.UpdateTransactionMethodIDToNull(uid, nil, models.BankAcc)
	go db.RedisDB.Del(context.TODO(), ("ba/" + uid))
	c.JSON(http.StatusOK, gin.H{"message": "Bank accounts deleted successfully by user id."})
}
