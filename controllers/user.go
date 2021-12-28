package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

var (
	errFailedRegister = errors.New("failed to register")
)

func (u *UserController) Register(c *gin.Context) {
	//TODO: Create new user model
	// email confirmation
	// create base response

	c.JSON(http.StatusCreated, gin.H{"message": "registered successfully"})
}

//TODO: Password Forgotten
// Change password
// Delete
