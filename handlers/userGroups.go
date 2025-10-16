package handlers

import (
	"api-server/orm"
	"context"
	"net/http"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CreateUserGroup godoc
//
//	@Summary		Creates a new empty User-Group
//	@Description	Creates a new casbin group (corresponds to entry g, nullUser, <name-of-group>)
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"Group name"
//	@Success		201		{object}	map[string]string	"Empty group created successfully!"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/create-ugroup [post]
func CreateUserGroup(ctx *gin.Context) {
	// Get group name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "group name cannot be empty"})

		return
	}

	err := auth.CreateUserGroup(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to create user group")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user group"})

		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Group created successfully!",
		"group":   name,
	})
}

// RemoveUserGroup godoc
//
//	@Summary		Removes a User-Group
//	@Description	Removes a casbin group with all its policies and group definitions
//	@Description	Throws error, if group does not exist or is enclaveAdmin
//
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"Group name"
//	@Success		201		{object}	map[string]string	"Group removed successfully!"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/remove-ugroup [post]
func RemoveUserGroup(ctx *gin.Context) {
	// Get group name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "group name cannot be empty"})

		return
	}

	err := auth.RemoveUserGroup(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove user group")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove user group"})

		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Group removed successfully!",
		"group":   name,
	})
}

// GetUserGroups godoc
//
//	@Summary		Get User-Groups
//	@Description	Gets all User-Groups its policies and group definitions
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		201		body        []auth.UserGroup	"user groups"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/ugroups [get]
func GetUserGroups(ctx *gin.Context) {
	users, err := auth.GetUserGroups()
	if err != nil {
		log.Error().Err(err).Msg("failed to get user groups")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user groups"})

		return
	}
	// Success response
	ctx.JSON(http.StatusOK, users)
}

// AddToUserGroup godoc
//
//	@Summary		Add a User to a User-Group
//	@Description	Adds a user to one or more user-groups
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			AddToGroupBody	body		AddToGroupBody		true	"Added successfully!"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/add-to-ugroup [post]
func AddToUserGroup(ctx *gin.Context) {
	var body AddToGroupBody

	// Try to convert the provided body to AddToGroupBody struct
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})

		return
	}

	err := auth.AddUserToGroup(body.Username, body.Groups...)
	if err != nil {
		log.Error().Err(err).Msg("failed to add user to group")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add user to group"})

		return
	}

	// Print the values (as requested)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Added successfully!",
	})
}

// RemoveFromUserGroup godoc
//
//	@Summary		Removes a User from a User-Group
//	@Description	Removes a user from one or more user-groups
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			AddToGroupBody	body		AddToGroupBody		true	"Add user to groups"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/remove-from-ugroup [post]
func RemoveFromUserGroup(ctx *gin.Context) {
	var body AddToGroupBody

	// Try to convert the provided body to AddToGroupBody struct
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})

		return
	}

	err := auth.RemoveUserFromGroup(body.Username, body.Groups...)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove user from group")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove user from group"})

		return
	}

	// Print the values (as requested)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "User successfully removed from group!",
	})
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
	user, _ := gorm.G[orm.User](
		orm.DB,
	).Where(&orm.User{Username: name}).
		First(context.Background())

	orm.DB.Delete(&orm.Auth_Basic{}, "id = ?", user.ID)
	orm.DB.Delete(&orm.User{}, "id = ?", user.ID)

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User removed successfully!",
		"group":   name,
	})
}

// GetGroupsOfUser godoc
//
//	@Summary		Get Groups of a User
//	@Description	Returns all groups a user belongs to
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"User name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/groups-of [post]
func GetGroupsOfUser(ctx *gin.Context) {
	// Get group name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})

		return
	}

	groups, err := auth.GetGroupsForUser(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to get groups of user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get groups of user"})

		return
	}
	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"user":   name,
		"groups": groups,
	})
}

// GetUsersOfGroup godoc
//
//	@Summary		Get Users of a Group
//	@Description	Returns all users that belong to a specific group
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"User name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/users-of [post]
func GetUsersOfGroup(ctx *gin.Context) {
	// Get group name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "GroupName cannot be empty"})

		return
	}

	users, err := auth.GetUserGroup(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to get users of group")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users of group"})

		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"group": name,
		"users": users,
	})
}
