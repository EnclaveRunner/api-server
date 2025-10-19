package handlers

import (
	"api-server/orm"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func GetUser(ctx *gin.Context) {
	var body UserBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Error().Err(err).Msg("Failed to process request body")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			err.Error(),
		})

		return
	}

	uuidParsed, err := uuid.Parse(body.ID)
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID format")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			"Invalid UUID format for ID",
		})

		return
	}

	user, err := orm.GetUserByID(uuidParsed)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			ctx.Status(http.StatusNotFound)

			return
		}

		log.Error().Err(err).Msg("Failed to query user from database")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	ctx.JSON(http.StatusOK, user)
}

func ListUsers(ctx *gin.Context) {
	users, err := orm.ListAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("Failed to query users from database")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	ctx.JSON(http.StatusOK, users)
}

func CreateUser(ctx *gin.Context) {
	var body UserCreateBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Error().Err(err).Msg("Failed to process request body")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			err.Error(),
		})

		return
	}

	user, err := orm.CreateUser(body.Username, body.Password)
	if err != nil {
		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			ctx.JSON(http.StatusConflict, &ResponseError{
				err.Error(),
			})

			return
		}

		log.Error().Err(err).Msg("Failed to create user in database")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	log.Info().
		Str("userID", user.ID.String()).
		Str("userName", user.Username).
		Msg("Created new user")

	ctx.JSON(http.StatusCreated, user)
}

func PatchUser(ctx *gin.Context) {
	var body UserUpdateBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Error().Err(err).Msg("Failed to process request body")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			err.Error(),
		})

		return
	}

	uuidParsed, err := uuid.Parse(body.ID)
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID format")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			"Invalid UUID format for ID",
		})

		return
	}

	var newUsername *string
	if body.NewUsername != "" {
		newUsername = &body.NewUsername
	}

	var newPassword *string
	if body.NewPassword != "" {
		newPassword = &body.NewPassword
	}

	user, err := orm.PatchUser(uuidParsed, newUsername, newPassword)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			ctx.Status(http.StatusNotFound)

			return
		}

		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			ctx.JSON(http.StatusConflict, &ResponseError{
				err.Error(),
			})

			return
		}

		log.Error().Err(err).Msg("Failed to update user in database")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	ctx.JSON(http.StatusOK, user)
}

func DeleteUser(ctx *gin.Context) {
	var body UserBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Error().Err(err).Msg("Failed to process request body")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			err.Error(),
		})

		return
	}

	uuidParsed, err := uuid.Parse(body.ID)
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID format")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			"Invalid UUID format for ID",
		})

		return
	}

	user, err := orm.DeleteUserByID(uuidParsed)
	var errNotFound *orm.NotFoundError
	if err != nil {
		if errors.As(err, &errNotFound) {
			ctx.Status(http.StatusNotFound)

			return
		}

		log.Error().Err(err).Msg("Failed to delete user from database")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	ctx.JSON(http.StatusOK, user)
}

func GetMe(ctx *gin.Context) {
	userID, _, _ := ctx.Request.BasicAuth()
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID format in Basic Auth")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			"Invalid UUID format in Basic Auth",
		})

		return
	}

	user, err := orm.GetUserByID(userIDParsed)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			ctx.Status(http.StatusNotFound)

			return
		}

		log.Error().Err(err).Msg("Failed to query user from database")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	ctx.JSON(http.StatusOK, user)
}

func UpdateMe(ctx *gin.Context) {
	var body MeUpdateBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Error().Err(err).Msg("Failed to process request body")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			err.Error(),
		})

		return
	}

	userID, _, _ := ctx.Request.BasicAuth()
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID format in Basic Auth")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			"Invalid UUID format in Basic Auth",
		})

		return
	}

	var newUsername *string
	if body.NewUsername != "" {
		newUsername = &body.NewUsername
	}

	var newPassword *string
	if body.NewPassword != "" {
		newPassword = &body.NewPassword
	}

	user, err := orm.PatchUser(userIDParsed, newUsername, newPassword)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			log.Error().Err(err).Msg("Authenticated user not found in database")
			ctx.Status(http.StatusInternalServerError)

			return
		}

		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			ctx.JSON(http.StatusConflict, &ResponseError{
				err.Error(),
			})

			return
		}

		log.Error().Err(err).Msg("Failed to update user in database")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	ctx.JSON(http.StatusOK, user)
}

func DeleteMe(ctx *gin.Context) {
	userID, _, _ := ctx.Request.BasicAuth()
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		log.Error().Err(err).Msg("Invalid UUID format in Basic Auth")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			"Invalid UUID format in Basic Auth",
		})

		return
	}

	user, err := orm.DeleteUserByID(userIDParsed)
	var errNotFound *orm.NotFoundError
	if err != nil {
		if errors.As(err, &errNotFound) {
			ctx.Status(http.StatusNotFound)

			return
		}

		log.Error().Err(err).Msg("Failed to delete user from database")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	ctx.JSON(http.StatusOK, user)
}
