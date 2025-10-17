package handlers

import (
	"api-server/orm"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

	user, err := gorm.G[orm.User](orm.DB).
		Where(&orm.User{ID: uuidParsed}).
		First(context.Background())

	if err != nil {
		if err == gorm.ErrRecordNotFound {
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
	users, err := gorm.G[orm.User](orm.DB).Find(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to query users from database")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func CreateUser(ctx *gin.Context) {
	var body UserCreateBody

	// Try to convert the provided body to UserCreateBody struct
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Error().Err(err).Msg("Failed to process request body")
		ctx.JSON(http.StatusBadRequest, &ResponseError{
			err.Error(),
		})

		return
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), orm.HashCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	foundUser, err := gorm.G[orm.User](orm.DB).
		Where(&orm.User{Username: body.Username}).
		Count(context.Background(), "*")

	if err != nil {
		log.Error().Err(err).Msg("Failed to query user from database")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if foundUser > 0 {
		log.Error().Msg("User with given username already exists")
		ctx.JSON(http.StatusConflict, &ResponseError{
			"User with given username already exists",
		})
		return
	}

	orm.DB.Save(&orm.User{Username: body.Username})
	createdUser, err := gorm.G[orm.User](
		orm.DB,
	).Where(&orm.User{Username: body.Username}).
		First(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to query created user from database")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	err = orm.DB.Transaction(func(tx *gorm.DB) error {
		// Create user
		if err := tx.Create(&orm.User{Username: body.Username}).Error; err != nil {
			return err
		}

		// Retrieve created user
		if err := tx.Where(&orm.User{Username: body.Username}).First(&createdUser).Error; err != nil {
			return err
		}

		// Create auth record
		if err := tx.Create(&orm.Auth_Basic{UserID: createdUser.ID, Password: hash}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusCreated)
}

func UpdateUser(ctx *gin.Context) {
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

	user, err := gorm.G[orm.User](orm.DB).
		Where(&orm.User{ID: uuidParsed}).
		First(context.Background())

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.Status(http.StatusNotFound)
			return
		}

		log.Error().Err(err).Msg("Failed to query user from database")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	err = orm.DB.Transaction(func(tx *gorm.DB) error {
		if body.NewUsername != "" {
			if err := tx.Save(&orm.User{
				ID:       user.ID,
				Username: body.NewUsername,
			}).Error; err != nil {
				return err
			}
		}

		if body.NewPassword != "" {
			// hash password
			hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), orm.HashCost)
			if err != nil {
				return err
			}

			if err := tx.Save(&orm.Auth_Basic{
				UserID:   user.ID,
				Password: hash,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to update user")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
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

	gorm.G[orm.User](orm.DB).Where(&orm.User{
		ID: uuidParsed,
	}).Delete(context.Background())
}

func GetMe(ctx *gin.Context) {

}
func UpdateMe(ctx *gin.Context) {}
func DeleteMe(ctx *gin.Context) {}
