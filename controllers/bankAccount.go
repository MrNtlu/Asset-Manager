package controllers

import (
	"asset_backend/db"
	"asset_backend/models"
	"asset_backend/requests"
	"context"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/vmihailenco/msgpack/v5"
)

type BankAccountController struct {
	Database *db.MongoDB
}

func NewBankAccountController(mongoDB *db.MongoDB) BankAccountController {
	return BankAccountController{
		Database: mongoDB,
	}
}

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
// @Failure 403 {string} string
// @Failure 500 {string} string
// @Router /ba [post]
func (ba *BankAccountController) CreateBankAccount(c *gin.Context) {
	var data requests.BankAccountCreate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(ba.Database)
	isPremium := userModel.IsUserPremium(uid)

	bankAccModel := models.NewBankAccountModel(ba.Database)
	if !isPremium && bankAccModel.GetUserBankAccountCount(uid) >= 2 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errBankAccountPremium,
		})

		return
	}

	var (
		createdBankAccount models.BankAccount
		err                error
	)

	if createdBankAccount, err = bankAccModel.CreateBankAccount(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	ba.clearCache(uid)

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

	var (
		cacheKey     = "ba/" + uid
		bankAccounts []models.BankAccount
	)

	result, err := db.RedisDB.Get(context.TODO(), cacheKey).Result()
	if err != nil || result == "" {
		bankAccModel := models.NewBankAccountModel(ba.Database)
		bankAccounts, err = bankAccModel.GetBankAccountsByUserID(uid)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		marshalBankAccounts, _ := msgpack.Marshal(bankAccounts)
		go db.RedisDB.Set(context.TODO(), cacheKey, marshalBankAccounts, db.RedisLExpire)
	} else if err := msgpack.Unmarshal([]byte(result), &bankAccounts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": bankAccounts})
}

// BankAccounts Stats
// @Summary Get Bank Accounts Stats
// @Description Returns bank accounts stats
// @Tags bankaccount
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.TransactionTotal
// @Failure 500 {string} string
// @Router /ba/stats [get]
func (ba *BankAccountController) GetBankAccountStatistics(c *gin.Context) {
	var data requests.TransactionMethod
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	transactionModel := models.NewTransactionModel(ba.Database)

	bankAccountStats, err := transactionModel.GetMethodStatistics(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": bankAccountStats})
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

	bankAccModel := models.NewBankAccountModel(ba.Database)

	bankAccount, err := bankAccModel.GetBankAccountByID(data.ID)
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

	if updatedBankAccount, err = bankAccModel.UpdateBankAccount(data, bankAccount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	ba.clearCache(uid)

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
	bankAccModel := models.NewBankAccountModel(ba.Database)

	isDeleted, err := bankAccModel.DeleteBankAccountByBAID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		ba.clearCache(uid)
		transactionModel := models.NewTransactionModel(ba.Database)

		go transactionModel.UpdateTransactionMethodIDToNull(uid, &data.ID, models.BankAcc)
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

	bankAccModel := models.NewBankAccountModel(ba.Database)
	if err := bankAccModel.DeleteAllBankAccountsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	transactionModel := models.NewTransactionModel(ba.Database)
	go transactionModel.UpdateTransactionMethodIDToNull(uid, nil, models.BankAcc)
	ba.clearCache(uid)
	c.JSON(http.StatusOK, gin.H{"message": "Bank accounts deleted successfully by user id."})
}

func (ba *BankAccountController) clearCache(uid string) {
	go db.RedisDB.Del(context.TODO(), ("ba/" + uid))
}
