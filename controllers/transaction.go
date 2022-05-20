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

type TransactionController struct{}

var (
	//TODO: Limit per day (by transaction date)
	//TODO: Also check on update
	errTransactionMethodUnauthorized = "Unauthorized method access. You're not authorized for this method."
	errTransactionPremium            = "Free members can add up to x transactions per day, you can get premium membership for unlimited access."
	errNoTransaction                 = "Couldn't find transaction."
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
	// isPremium := models.IsUserPremium(uid)
	//TODO: Add premium check

	switch *data.TransactionMethod.Type {
	case 0:
		bankAccount, err := models.GetBankAccountByID(data.TransactionMethod.MethodID)
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
		creditCard, err := models.GetCardByID(data.TransactionMethod.MethodID)
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

	var (
		createdTransaction models.Transaction
		err                error
	)

	if createdTransaction, err = models.CreateTransaction(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go db.RedisDB.Del(context.TODO(), ("transaction/" + uid))

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdTransaction})
}

// Transaction Calendar Count
// @Summary Get Number of Transactions per Day for Calendar
// @Description Returns number of transactions by year and month
// @Tags transaction
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.TransactionCalendarCount
// @Failure 500 {string} string
// @Router /transaction/calendar [get]
func (t *TransactionController) GetCalendarTransactionCount(c *gin.Context) {
	var data requests.TransactionCalendar
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	transactionCalendarCounts, err := models.GetCalendarTransactionCount(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": transactionCalendarCounts})
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

	if (data.StartDate != nil && data.EndDate == nil) || (data.StartDate == nil && data.EndDate != nil) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Please select date range correctly.",
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	transactions, pagination, err := models.GetTransactionsByUserIDAndFilterSort(uid, data)
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

	transaction, err := models.GetTransactionByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if transaction.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": errNoTransaction})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != transaction.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	if data.TransactionMethod != nil {
		switch *data.TransactionMethod.Type {
		case 0:
			bankAccount, err := models.GetBankAccountByID(data.TransactionMethod.MethodID)
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
			creditCard, err := models.GetCardByID(data.TransactionMethod.MethodID)
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
	if updatedTransaction, err = models.UpdateTransaction(data, transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go db.RedisDB.Del(context.TODO(), ("transaction/" + uid))

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
	isDeleted, err := models.DeleteTransactionByTransactionID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if isDeleted {
		go db.RedisDB.Del(context.TODO(), ("transaction/" + uid))
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
	if err := models.DeleteAllTransactionsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	go db.RedisDB.Del(context.TODO(), ("transaction/" + uid))
	c.JSON(http.StatusOK, gin.H{"message": "Transactions deleted successfully by user id."})
}
