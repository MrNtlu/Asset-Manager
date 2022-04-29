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

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuth2Controller struct{}

var (
	googleOauthConfig *oauth2.Config

	oauthStateString = "kantan-login"

	errFailedLogin = "failed to login"
)

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
		json.NewDecoder(response.Body).Decode(&googleToken)
		if googleToken.Email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": errFailedLogin, "code": http.StatusUnauthorized})
			return
		}

		var user models.User
		user, _ = models.FindUserByEmail(googleToken.Email)
		if user.EmailAddress == "" {
			oAuthUser, err := models.CreateOAuthUser(googleToken.Email)
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

		c.SetCookie("jwt", token, 259200, "/", os.Getenv("BASE_URI"), true, true)
		c.JSON(http.StatusOK, gin.H{"access_token": token})
	}
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
		json.Unmarshal(content, &authGoogle)

		var user models.User
		user, _ = models.FindUserByEmail(authGoogle.Email)
		if user.EmailAddress == "" {
			oAuthUser, err := models.CreateOAuthUser(authGoogle.Email)
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

		c.SetCookie("access_token", token, 259200, "/", os.Getenv("BASE_URI"), true, true)
		c.JSON(http.StatusOK, gin.H{"access_token": token})
	}
}

func getUserInfo(state string, code string) ([]byte, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	return contents, nil
}
