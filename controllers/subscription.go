package controllers

import (
	"asset_backend/db"
	"asset_backend/helpers"
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"net/http"
	"sort"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/vmihailenco/msgpack/v5"
)

type SubscriptionController struct {
	Database *db.MongoDB
}

func NewSubscriptionController(mongoDB *db.MongoDB) SubscriptionController {
	return SubscriptionController{
		Database: mongoDB,
	}
}

var (
	errSubscriptionNotFound            = "Subscription not found."
	errSubscriptionInviteSelf          = "You cannot invite yourself."
	errUnauthorizedCreditCard          = "Unauthorized credit card access. You're not the owner of this credit card."
	errAlreadyShared                   = "This user already has access to subscription."
	errSubscriptionPremium             = "Free members can add up to 5 subscriptions, you can get premium membership for unlimited access."
	errSubscriptionNotificationPremium = "You should be premium user for this feature."
)

// Create Subscription
// @Summary Create Subscription
// @Description Creates subscription
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param subscription body requests.Subscription true "Subscription Create"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} responses.Subscription
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription [post]
func (s *SubscriptionController) CreateSubscription(c *gin.Context) {
	var data requests.Subscription
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(s.Database)
	isPremium := userModel.IsUserPremium(uid)

	subscriptionModel := models.NewSubscriptionModel(s.Database)
	if !isPremium && subscriptionModel.GetUserSubscriptionCount(uid) >= 5 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errSubscriptionPremium,
		})

		return
	}

	if data.NotificationTime != nil && !isPremium {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errSubscriptionNotificationPremium,
		})

		return
	}

	if data.CardID != nil {
		cardModel := models.NewCardModel(s.Database)

		creditCard, tempErr := cardModel.GetCardByID(*data.CardID)
		if tempErr == nil && creditCard.UserID != uid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errUnauthorizedCreditCard,
			})

			return
		} else if tempErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": tempErr.Error(),
			})

			return
		}
	}

	var (
		createdSubscription responses.Subscription
		err                 error
	)

	if createdSubscription, err = subscriptionModel.CreateSubscription(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	go db.RedisDB.Del(context.TODO(), ("subscription/" + uid))

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdSubscription})
}

// Invite Subscription To User
// @Summary Sents invitation to another user for access to subscription.
// @Description Invites another user by email to subscription, if accepted they can view it
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param subscriptioninvite body requests.SubscriptionInvite true "SubscriptionInvite"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription/invite [post]
func (s *SubscriptionController) InviteSubscriptionToUser(c *gin.Context) {
	var data requests.SubscriptionInvite
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(s.Database)

	user, err := userModel.FindUserByEmail(data.InvitedUserMail)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Invitation sent. Please ask them to check their invitation & accept it.",
		})

		return
	}

	if user.ID.Hex() == uid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errSubscriptionInviteSelf,
		})

		return
	}

	subscriptionModel := models.NewSubscriptionModel(s.Database)

	subscription, err := subscriptionModel.GetSubscriptionByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if subscription.UserID != uid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": ErrUnauthorized,
		})

		return
	}

	for _, userID := range subscription.SharedUsers {
		if userID == user.ID.Hex() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errAlreadyShared,
			})

			return
		}
	}

	if err := subscriptionModel.InviteSubscriptionToUser(uid, user.ID.Hex(), data.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	go helpers.SendNotification(user.FCMToken, "Subscription Invitation", "Subscription share invitation received.")

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation sent. Please ask them to check their invitation & accept it.",
	})
}

// Cancel Subscription Invitation
// @Summary Cancel Subscription Invitation
// @Description Cancels subscription invitation
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param ID query requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription/cancel [post]
func (s *SubscriptionController) CancelSubscriptionInvitation(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionModel := models.NewSubscriptionModel(s.Database)

	if err := subscriptionModel.CancelSubscriptionInvitation(data.ID, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation cancelled successfully.",
	})
}

// Shared Subscriptions By User ID
// @Summary Get Shared Subscriptions by User ID
// @Description Returns shared subscriptions by user id
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param ID query requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} models.Subscription
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription/shared [get]
func (s *SubscriptionController) GetSharedSubscriptionsByUserID(c *gin.Context) {
	var data requests.SubscriptionSort
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionModel := models.NewSubscriptionModel(s.Database)

	sharedSubscriptions, err := subscriptionModel.GetSubscriptionsByUserID(uid, data, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": sharedSubscriptions})
}

// Handles Subscription Share Invitation
// @Summary Handles subscription invitation response
// @Description Invited users can accept/deny invitation
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param subscriptioninvitation body requests.SubscriptionInvitation true "SubscriptionInvitation"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription/invitation [post]
func (s *SubscriptionController) HandleSubscriptionInvitation(c *gin.Context) {
	var data requests.SubscriptionInvitation
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionModel := models.NewSubscriptionModel(s.Database)

	if err := subscriptionModel.HandleSubscriptionInvitation(data.ID, uid, *data.IsAccepted); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful.",
	})
}

// Subscriptions By Card
// @Summary Get Subscriptions by Card ID
// @Description Returns subscriptions by card id
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param ID query requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} models.Subscription
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription/card [get]
func (s *SubscriptionController) GetSubscriptionsByCardID(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionModel := models.NewSubscriptionModel(s.Database)

	subscriptions, err := subscriptionModel.GetSubscriptionsByCardID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": subscriptions})
}

// Subscriptions and Stats
// @Summary Subscriptions and Stats by User ID
// @Description Returns subscriptions and stats by user id
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param subscriptionsort query requests.SubscriptionSort true "Subscription Sort"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.SubscriptionAndStats
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscription [get]
func (s *SubscriptionController) GetSubscriptionsAndStatsByUserID(c *gin.Context) {
	var data requests.SubscriptionSort
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	var (
		cacheKey             = "subscription/" + uid
		subscriptionAndStats responses.SubscriptionAndStats
	)

	result, err := db.RedisDB.Get(context.TODO(), cacheKey).Result()
	if err != nil || result == "" {
		subscriptionModel := models.NewSubscriptionModel(s.Database)

		subscriptions, err := subscriptionModel.GetSubscriptionsByUserID(uid, data, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		subscriptionStats, err := subscriptionModel.GetSubscriptionStatisticsByUserID(uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		subscriptionAndStats = responses.SubscriptionAndStats{
			Data:  subscriptions,
			Stats: subscriptionStats,
		}

		marshalSubscriptionAndStats, _ := msgpack.Marshal(subscriptionAndStats)
		go db.RedisDB.Set(context.TODO(), cacheKey, marshalSubscriptionAndStats, db.RedisLExpire)
	} else {
		msgpack.Unmarshal([]byte(result), &subscriptionAndStats)
		sort.Slice(subscriptionAndStats.Data, func(i, j int) bool {
			switch data.Sort {
			case "name":
				if data.SortType == 1 {
					return subscriptionAndStats.Data[i].Name < subscriptionAndStats.Data[j].Name
				}

				return subscriptionAndStats.Data[i].Name > subscriptionAndStats.Data[j].Name
			case "currency":
				if data.SortType == 1 {
					return subscriptionAndStats.Data[i].Currency < subscriptionAndStats.Data[j].Currency
				}

				return subscriptionAndStats.Data[i].Currency > subscriptionAndStats.Data[j].Currency
			case "price":
				if data.SortType == 1 {
					return subscriptionAndStats.Data[i].Price > subscriptionAndStats.Data[j].Price
				}

				return subscriptionAndStats.Data[i].Price < subscriptionAndStats.Data[j].Price
			case "date":
				if data.SortType == 1 {
					return subscriptionAndStats.Data[i].NextBillDate.After(subscriptionAndStats.Data[j].NextBillDate)
				}

				return subscriptionAndStats.Data[i].NextBillDate.Before(subscriptionAndStats.Data[j].NextBillDate)
			default:
				return true
			}
		})
	}

	c.JSON(http.StatusOK, subscriptionAndStats)
}

// Subscription Details
// @Summary Get Subscription Details
// @Description Returns subscription details
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param ID query requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.SubscriptionDetails
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /subscription/details [get]
func (s *SubscriptionController) GetSubscriptionDetails(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionModel := models.NewSubscriptionModel(s.Database)

	subscription, err := subscriptionModel.GetSubscriptionDetails(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if subscription.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": errSubscriptionNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": subscription})
}

// Subscription Statistics
// @Summary Get Subscription Statistics by User ID
// @Description Returns subscription statistics
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.SubscriptionStatistics
// @Failure 500 {string} string
// @Router /subscription/stats [get]
func (s *SubscriptionController) GetSubscriptionStatisticsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionModel := models.NewSubscriptionModel(s.Database)

	subscriptionStats, err := subscriptionModel.GetSubscriptionStatisticsByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully fetched.", "data": subscriptionStats})
}

// Update Subscription
// @Summary Update Subscription
// @Description Updates subscription
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param subscriptionupdate body requests.SubscriptionUpdate true "Subscription Update"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.Subscription
// @Failure 400 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Couldn't find user"
// @Failure 500 {string} string
// @Router /subscription [put]
func (s *SubscriptionController) UpdateSubscription(c *gin.Context) {
	var data requests.SubscriptionUpdate
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	subscriptionModel := models.NewSubscriptionModel(s.Database)

	subscription, err := subscriptionModel.GetSubscriptionByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if subscription.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": errSubscriptionNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != subscription.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	var updatedSubscription responses.Subscription

	if data.NotificationTime != nil {
		userModel := models.NewUserModel(s.Database)
		if isPremium := userModel.IsUserPremium(uid); !isPremium {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errSubscriptionNotificationPremium,
			})

			return
		}
	}

	if updatedSubscription, err = subscriptionModel.UpdateSubscription(data, subscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	go db.RedisDB.Del(context.TODO(), ("subscription/" + uid))

	c.JSON(http.StatusOK, gin.H{"message": "Subscription updated.", "data": updatedSubscription})
}

// Delete Subscription By ID
// @Summary Delete subscription by subscription id
// @Description Deletes subscription by id
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Param ID body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /subscription [delete]
func (s *SubscriptionController) DeleteSubscriptionBySubscriptionID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionModel := models.NewSubscriptionModel(s.Database)

	isDeleted, err := subscriptionModel.DeleteSubscriptionBySubscriptionID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		go db.RedisDB.Del(context.TODO(), ("subscription/" + uid))
		c.JSON(http.StatusOK, gin.H{"message": "Subscription deleted successfully."})

		return
	}

	c.JSON(http.StatusOK, gin.H{"error": ErrUnauthorized})
}

// Delete All Subscriptions
// @Summary Delete all subscriptions by user id
// @Description Deletes all subscriptions by user id
// @Tags subscription
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /subscription/all [delete]
func (s *SubscriptionController) DeleteAllSubscriptionsByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	subscriptionModel := models.NewSubscriptionModel(s.Database)

	if err := subscriptionModel.DeleteAllSubscriptionsByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	go db.RedisDB.Del(context.TODO(), ("subscription/" + uid))

	c.JSON(http.StatusOK, gin.H{"message": "Subscriptions deleted successfully by user id."})
}
