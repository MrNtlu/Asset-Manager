package helpers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/utils"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var identityKey = "id"

func SetupJWTHandler() *jwt.GinJWTMiddleware {
	port := os.Getenv("PORT")
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	if port == "" {
		port = "8080"
	}

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "asset manager",
		Key:         []byte(os.Getenv("JWT_SECRET_KEY")),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour * 8760,
		IdentityKey: identityKey,
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var data requests.Login
			if err := c.Bind(&data); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			user, err := models.FindUserByEmail(data.EmailAddress)
			if err != nil {
				return "", jwt.ErrFailedAuthentication
			}

			if user.Password == "" {
				return "", errors.New("password is empty")
			}

			if err := utils.CheckPassword([]byte(user.Password), []byte(data.Password)); err != nil {
				return "", jwt.ErrFailedAuthentication
			}

			return user, nil
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if user, ok := data.(models.User); ok {
				return jwt.MapClaims{
					identityKey: user.ID,
				}
			}
			return jwt.MapClaims{}
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.SetCookie("access_token", token, 3600, "/", os.Getenv("BASE_URI"), true, true)
			c.JSON(http.StatusOK, gin.H{"access_token": token})
		},
		RefreshResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.SetCookie("access_token", token, 3600, "/", os.Getenv("BASE_URI"), true, true)
			c.JSON(http.StatusOK, gin.H{"access_token": token})
		},
		TokenLookup:    "header: Authorization, cookie: jwt_token",
		TimeFunc:       time.Now,
		SendCookie:     true,
		SecureCookie:   false,       //non HTTPS dev environments
		CookieHTTPOnly: true,        // JS can't modify
		CookieName:     "jwt_token", // default jwt
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	return authMiddleware
}
