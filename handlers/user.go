package handlers

import (
	"api-server/orm"
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
		if _, ok := err.(*orm.NotFoundError); ok {
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

func PutUser(ctx *gin.Context) {
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
		if _, ok := err.(*orm.ConflictError); ok {
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
		if _, ok := err.(*orm.NotFoundError); ok {
			ctx.Status(http.StatusNotFound)
			return
		}

		if _, ok := err.(*orm.ConflictError); ok {
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
	if err != nil {
		if _, ok := err.(*orm.NotFoundError); ok {
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

}
func UpdateMe(ctx *gin.Context) {}
func DeleteMe(ctx *gin.Context) {}
