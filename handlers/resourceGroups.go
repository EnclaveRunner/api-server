package handlers

import (
	"net/http"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CreateResourceGroup godoc
//
//	@Summary		Creates a new Resource-Group
//	@Description	Adds casbin entry g, nullResource, [name]
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"User name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/create-rgroup [post]
func CreateResourceGroup(ctx *gin.Context) {
	// Get group name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "GroupName cannot be empty"})

		return
	}

	err := auth.CreateResourceGroup(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to create resource group")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create resource group"})

		return
	}
	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Resource-Group created successfully!",
	})
}

// RemoveResourceGroup godoc
//
//	@Summary		Remove a Resource-Group
//	@Description	Removes casbin entry g, nullResource, [name]
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"User name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/remove-rgroup [post]
func RemoveResourceGroup(ctx *gin.Context) {
	// Get group name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "GroupName cannot be empty"})

		return
	}

	err := auth.RemoveResourceGroup(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove resource group")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove resource group"})

		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Resource-Group removed successfully!",
	})
}

// GetResourceGroups godoc
//
//	@Summary		Get all Resource-Groups
//	@Description	Lists all resource-groups
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"User name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/rgroups [get]
func GetResourceGroups(ctx *gin.Context) {
	rgroups, err := auth.GetResourceGroups()
	if err != nil {
		log.Error().Err(err).Msg("failed to remove resource group")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove resource group"})

		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"groups": rgroups,
	})
}

// AddToResourceGroup godoc
//
//	@Summary		Add Resource to Resource-Group
//	@Description	Adds resource to specified resource-group
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"User name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/add-to-rgroup [post]
func AddToResourceGroup(ctx *gin.Context) {
	var body AddToRGroupBody

	// Try to convert the provided body to AddToGroupBody struct
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})

		return
	}

	err := auth.AddResourceToGroup(body.Resource, body.Groups...)
	if err != nil {
		log.Error().Err(err).Msg("failed to add resource to resource group")
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to add resource to resource group"},
		)

		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Resource added to Resource-Group successfully!",
	})
}
