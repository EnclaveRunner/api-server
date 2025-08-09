package handlers

import (
	"enclave-backend/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Demo godoc
//
//	@Summary		Demo endpoint
//	@Description	A simple demo endpoint to show API functionality
//	@Tags			demo
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				false	"Name to greet"
//	@Success		200		{object}	map[string]string	"message"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/demo [get]
func Demo(ctx *gin.Context) {
	name := ctx.DefaultQuery("param", "")
	if name == "error" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "name parameter is required"})

		return
	}

	if name == "not-found" {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "not found"})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Hello, " + name + "! This is a demo endpoint!"})
}

// IssueToken godoc
//
//	@Summary		Issue JWT token
//	@Description	Issues a JWT token for a given username and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			username	query		string	true	"Username"
//	@Param			password	query		string	true	"Password"
//	@Success		200			{object}	string	"JWT token"
//	@Failure		403			{object}	string	"invalid username or password"
//	@Router			/issue-token [get]
func IssueToken(ctx *gin.Context) {
	// Mock authentication - proper authentication should be implemented for production
	username := ctx.Query("username")
	password := ctx.Query("password")
	token, err := middleware.IssueJWT(ctx, username, password)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
