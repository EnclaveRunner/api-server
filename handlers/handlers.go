package handlers

import (
	"api-server/orm"
	"context"
	"net/http"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Ready godoc
//
//	@Summary		Health-Check Endpoint
//	@Description	Health-Check to see if the Api-Server is reachable / ready
//	@Tags			system
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	map[string]string	"{status: ready}"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/ready [get]
func Ready(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// RemoveUser godoc
//
//	@Summary		Removes a User entirely
//	@Description	Removes a user including all its group memberships
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"User name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/remove-user [post]
func RemoveUser(ctx *gin.Context) {
	// Get group name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})

		return
	}

	err := auth.RemoveUser(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove user"})

		return
	}

	// Remove user from User and Auth_Basic table
	user, err := gorm.G[orm.User](
		orm.DB,
	).Where(&orm.User{Username: name}).
		First(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("failed to find user in database")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find user in database"})

		return
	}

	orm.DB.Delete(&orm.Auth_Basic{}, "id = ?", user.ID)
	orm.DB.Delete(&orm.User{}, "id = ?", user.ID)

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User removed successfully!",
		"group":   name,
	})
}

// UpdateUser godoc
//
//	@Summary		Update a User's settings
//	@Description	Change username and/or password of a user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	map[string]string	"User created successfully!"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/ready [get]
func UpdateUser(ctx *gin.Context) {
	userID, _, _ := ctx.Request.BasicAuth()
	var body UserUpdateBody

	// Try to convert the provided body to UserCreateBody struct
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Fatal().Err(err).Msg("Failed to process request body")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})

		return
	}

	if body.NewPassword == "" ||
		body.NewUsername == "" {
		log.Fatal().Msg("Failed to process request body, request body invalid")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})

		return
	}

	user, err := gorm.G[orm.User](
		orm.DB,
	).Where(&orm.User{ID: uuid.MustParse(userID)}).
		First(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("User not with given ID not found ")
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})

		return
	}

	if body.NewUsername != "" {
		user, err := gorm.G[orm.User](
			orm.DB,
		).Where(&orm.User{ID: uuid.MustParse(userID)}).
			First(context.Background())
	}

	// Success response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User created successfully!",
	})
}
