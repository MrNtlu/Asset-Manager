package controllers

import (
	"asset_backend/models"
	"asset_backend/requests"
	"asset_backend/responses"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Timothylock/go-signin-with-apple/apple"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuth2Controller struct{}

var (
	googleOauthConfig *oauth2.Config

	oauthStateString = "kantan-login"

	errFailedLogin = "failed to login"

	errWrongLoginMethod = "Failed to login. This email is already registered with different login method."
)

const tokenExpiration = 259200

// OAuth2 Apple Login
// @Summary OAuth2 Apple Login
// @Description Gets user info from apple and creates/finds user and returns token
// @Tags oauth2
// @Accept application/json
// @Produce application/json
// @Success 200 {string} string "Token"
// @Failure 500 {string} string
// @Router /oauth/apple [post]
func (o *OAuth2Controller) OAuth2AppleLogin(jwt *jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data requests.AppleSignin
		if shouldReturn := bindJSONData(&data, c); shouldReturn {
			return
		}

		teamID := os.Getenv("TEAM_ID")
		clientID := os.Getenv("CLIENT_ID")
		keyID := os.Getenv("KEY_ID")
		secretKey := os.Getenv("SECRET_KEY")

		fmt.Println(secretKey)
		secret, err := apple.GenerateClientSecret(secretKey, teamID, clientID, keyID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"secret_key": secretKey,
			}).Error("Failed to generate secret key", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		appleClient := apple.New()

		if *data.IsRefresh {
			refreshRequest := apple.ValidationRefreshRequest{
				ClientID:     clientID,
				ClientSecret: secret,
				RefreshToken: data.Code,
			}

			var refreshResp apple.RefreshResponse

			err = appleClient.VerifyRefreshToken(context.Background(), refreshRequest, &refreshResp)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if refreshResp.Error != "" || refreshResp.AccessToken == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": (refreshResp.Error + " " + refreshResp.ErrorDescription)})
				return
			}

			var user models.User
			user, err = models.FindUserByRefreshToken(data.Code)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if !user.IsOAuthUser || (user.IsOAuthUser && user.OAuthType != 1) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": errWrongLoginMethod})
				return
			}

			token, _, err := jwt.TokenGenerator(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.SetCookie("jwt", token, tokenExpiration, "/", os.Getenv("BASE_URI"), true, true)
			c.JSON(http.StatusOK, gin.H{"access_token": token})

			return
		} else {
			req := apple.AppValidationTokenRequest{
				ClientID:     clientID,
				ClientSecret: secret,
				Code:         data.Code,
			}

			var resp apple.ValidationResponse

			err = appleClient.VerifyAppToken(context.Background(), req, &resp)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if resp.Error != "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": (resp.Error + " " + resp.ErrorDescription)})
				return
			}

			claim, err := apple.GetClaims(resp.IDToken)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			email := (*claim)["email"].(string)

			var user models.User
			user, _ = models.FindUserByEmail(email)
			if user.EmailAddress == "" {
				oAuthUser, err := models.CreateOAuthUser(email, &resp.RefreshToken, 1)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				user = *oAuthUser
			}

			if !user.IsOAuthUser || (user.IsOAuthUser && user.OAuthType != 1) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": errWrongLoginMethod})
				return
			}

			user.RefreshToken = &resp.RefreshToken
			if err := models.UpdateUser(user); err != nil {
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}

			token, _, err := jwt.TokenGenerator(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.SetCookie("jwt", token, tokenExpiration, "/", os.Getenv("BASE_URI"), true, true)
			c.JSON(http.StatusOK, gin.H{"access_token": token, "refresh_token": resp.RefreshToken})
		}
	}
}

// OAuth2 Google Login
// @Summary OAuth2 Google Login
// @Description Gets user info from google and creates/finds user and returns token
// @Tags oauth2
// @Accept application/json
// @Produce application/json
// @Success 200 {string} string "Token"
// @Failure 500 {string} string
// @Router /oauth/google [post]
func (o *OAuth2Controller) OAuth2GoogleLogin(jwt *jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data requests.GoogleLogin
		if shouldReturn := bindJSONData(&data, c); shouldReturn {
			return
		}

		response, err := http.Get("https://www.googleapis.com/oauth2/v3/tokeninfo?access_token=" + data.Token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errFailedLogin,
			})

			return
		}
		defer response.Body.Close()

		var googleToken responses.GoogleToken
		if err := json.NewDecoder(response.Body).Decode(&googleToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if googleToken.Email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": errFailedLogin, "code": http.StatusUnauthorized})
			return
		}

		var user models.User
		user, _ = models.FindUserByEmail(googleToken.Email)

		if user.EmailAddress == "" {
			oAuthUser, err := models.CreateOAuthUser(googleToken.Email, nil, 0)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			user = *oAuthUser
		}

		if !user.IsOAuthUser || (user.IsOAuthUser && user.OAuthType != 0) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errWrongLoginMethod})
			return
		}

		token, _, err := jwt.TokenGenerator(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.SetCookie("jwt", token, tokenExpiration, "/", os.Getenv("BASE_URI"), true, true)
		c.JSON(http.StatusOK, gin.H{"access_token": token})
	}
}

func SetOAuth2() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

func (o *OAuth2Controller) GoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(c.Writer, c.Request, url, http.StatusTemporaryRedirect)
}

// Google Callback
// @Summary Callback from Google OAuth
// @Description Callback from google auth
// @Tags oauth2
// @Accept application/json
// @Produce application/json
// @Success 200 {string} string "Token"
// @Failure 500 {string} string
// @Router /callback [get]
func (o *OAuth2Controller) GoogleCallback(jwt *jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		content, err := getUserInfo(c.Request.FormValue("state"), c.Request.FormValue("code"))
		if err != nil {
			http.Redirect(c.Writer, c.Request, "/", http.StatusTemporaryRedirect)
			return
		}

		type OAuth2Google struct {
			Email string `json:"email"`
		}

		var authGoogle OAuth2Google
		if err := json.Unmarshal(content, &authGoogle); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		var user models.User
		user, _ = models.FindUserByEmail(authGoogle.Email)

		if user.EmailAddress == "" {
			oAuthUser, err := models.CreateOAuthUser(authGoogle.Email, nil, 0)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			user = *oAuthUser
		}

		token, _, err := jwt.TokenGenerator(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.SetCookie("access_token", token, tokenExpiration, "/", os.Getenv("BASE_URI"), true, true)
		c.JSON(http.StatusOK, gin.H{"access_token": token})
	}
}

func getUserInfo(state string, code string) ([]byte, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}

	return contents, nil
}
