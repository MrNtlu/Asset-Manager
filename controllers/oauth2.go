package controllers

import (
	"asset_backend/models"
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

func (o *OAuth2Controller) GoogleCallback(jwt *jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		content, err := getUserInfo(c.Request.FormValue("state"), c.Request.FormValue("code"))
		if err != nil {
			fmt.Println(err.Error())
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
