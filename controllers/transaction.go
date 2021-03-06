package controllers

import (
	"asset_backend/db"
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/responses"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	Database *db.MongoDB
}

func NewTransactionController(mongoDB *db.MongoDB) TransactionController {
	return TransactionController{
		Database: mongoDB,
	}
}

var (
	errTransactionMethodUnauthorized = "Unauthorized method access. You're not authorized for this method."
	errTransactionPremium            = "Free members can add up to 10 transactions per day, you can get premium membership for unlimited access."
)

// Create Transaction
// @Summary Create Transaction
// @Description Creates transaction
// @Tags transaction
// @Accept application/json
// @Produce application/json
// @Param transaction body requests.TransactionCreate true "Transaction Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.Transaction
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /transaction [post]
func (t *TransactionController) CreateTransaction(c *gin.Context) {
	var data requests.TransactionCreate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(t.Database)
	isPremium := userModel.IsUserPremium(uid)

	transactionModel := models.NewTransactionModel(t.Database)
	if !isPremium && transactionModel.GetUserTransactionCountByTime(uid, data.TransactionDate) >= 10 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errTransactionPremium,
		})

		return
	}

	if data.TransactionMethod != nil {
		switch *data.TransactionMethod.Type {
		case 0:
			bankAccModel := models.NewBankAccountModel(t.Database)

			bankAccount, err := bankAccModel.GetBankAccountByID(data.TransactionMethod.MethodID)
			if err == nil && bankAccount.UserID != uid {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": errTransactionMethodUnauthorized,
				})

				return
			} else if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}
		case 1:
			cardModel := models.NewCardModel(t.Database)

			creditCard, err := cardModel.GetCardByID(data.TransactionMethod.MethodID)
			if err == nil && creditCard.UserID != uid {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": errTransactionMethodUnauthorized,
				})

				return
			} else if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}
		}
	}

	var (
		createdTransaction models.Transaction
		err                error
	)

	if createdTransaction, err = transactionModel.CreateTransaction(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdTransaction})
}

// Total Transaction By Interval
// @Summary Get Total Transaction Value by Interval
// @Description Returns total transaction value by interval
// @Tags transaction
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.TransactionTotal
// @Failure 500 {string} string
// @Router /transaction/total [get]
func (t *TransactionController) GetTotalTransactionByInterval(c *gin.Context) {
	var data requests.TransactionTotalInterval
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	transactionModel := models.NewTransactionModel(t.Database)

	transactionTotal, err := transactionModel.GetTotalTransactionByInterval(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": transactionTotal})
}

// Transaction Statistics
// @Summary Get Transaction Statistics by Interval
// @Description Returns transaction statistics by interval
// @Tags transaction
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.TransactionStats
// @Failure 500 {string} string
// @Router /transaction/stats [get]
func (t *TransactionController) GetTransactionStats(c *gin.Context) {
	var data requests.TransactionStatsInterval
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	transactionModel := models.NewTransactionModel(t.Database)

	transactionDailyStats, err := transactionModel.GetTransactionStats(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	categoryDistStats, err := transactionModel.GetTransactionCategoryDistribution(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	transactionStats := responses.TransactionStats{
		TransactionDailyStats:    transactionDailyStats,
		TransactionCategoryStats: categoryDistStats,
		TotalExpense:             transactionModel.GetTotalFromCategoryStats(categoryDistStats, false),
		TotalIncome:              transactionModel.GetTotalFromCategoryStats(categoryDistStats, true),
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": transactionStats})
}

// Transaction By User ID
// @Summary Get Transactions by User ID
// @Description Returns transactions by user id
// @Tags transaction
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} models.Transaction
// @Failure 500 {string} string
// @Router /transaction [get]
func (t *TransactionController) GetTransactionsByUserIDAndFilterSort(c *gin.Context) {
	var data requests.TransactionSortFilter
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	transactionModel := models.NewTransactionModel(t.Database)

	transactions, pagination, err := transactionModel.GetTransactionsByUserIDAndFilterSort(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions, "pagination": pagination})
}

// Update Transaction
// @Summary Update Transaction
// @Description Updates transaction
// @Tags transaction
// @Accept application/json
// @Produce application/json
// @Param transactionupdate body requests.TransactionUpdate true "Transaction Update"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.Transaction
// @Failure 400 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Couldn't find user"
// @Failure 500 {string} string
// @Router /transaction [put]
func (t *TransactionController) UpdateTransaction(c *gin.Context) {
	var data requests.TransactionUpdate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	transactionModel := models.NewTransactionModel(t.Database)

	transaction, err := transactionModel.GetTransactionByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != transaction.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	userModel := models.NewUserModel(t.Database)
	isPremium := userModel.IsUserPremium(uid)

	if data.TransactionDate != nil && !isPremium && transactionModel.GetUserTransactionCountByTime(uid, *data.TransactionDate) >= 10 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errTransactionPremium,
		})

		return
	}

	if data.TransactionMethod != nil {
		switch *data.TransactionMethod.Type {
		case 0:
			bankAccModel := models.NewBankAccountModel(t.Database)

			bankAccount, err := bankAccModel.GetBankAccountByID(data.TransactionMethod.MethodID)
			if err == nil && bankAccount.UserID != uid {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": errTransactionMethodUnauthorized,
				})

				return
			} else if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}
		case 1:
			cardModel := models.NewCardModel(t.Database)

			creditCard, err := cardModel.GetCardByID(data.TransactionMethod.MethodID)
			if err == nil && creditCard.UserID != uid {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": errTransactionMethodUnauthorized,
				})

				return
			} else if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}
		}
	}

	var updatedTransaction models.Transaction

	if updatedTransaction, err = transactionModel.UpdateTransaction(data, transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated.", "data": updatedTransaction})
}

// Delete Transaction By ID
// @Summary Delete transaction by transaction id
// @Description Deletes transaction by id
// @Tags transaction
// @Accept application/json
// @Produce application/json
// @Param ID body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /transaction [delete]
func (t *TransactionController) DeleteTransactionByTransactionID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	transactionModel := models.NewTransactionModel(t.Database)

	isDeleted, err := transactionModel.DeleteTransactionByTransactionID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": "Unauthorized delete."})
}

// Delete All Transactions
// @Summary Delete all transaction by user id
// @Description Deletes all transaction by user id
// @Tags transaction
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /transaction/all [delete]
func (t *TransactionController) DeleteAllTransactionsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	transactionModel := models.NewTransactionModel(t.Database)
	if err := transactionModel.DeleteAllTransactionsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transactions deleted successfully by user id."})
}
