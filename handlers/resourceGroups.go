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

// RemoveFromResourceGroup godoc
//
//	@Summary		Remove Resource from Resource-Group
//	@Description	Removes resource from specified resource-group
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"User name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/remove-from-rgroup [post]
func RemoveFromResourceGroup(ctx *gin.Context) {
	var body AddToRGroupBody

	// Try to convert the provided body to AddToGroupBody struct
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})

		return
	}

	err := auth.RemoveResourceFromGroup(body.Resource, body.Groups...)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove resource from resource group")
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to remove resource from resource group"},
		)

		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Resource removed from Resource-Group successfully!",
	})
}

// RemoveFromResourceGroup godoc
//
//	@Summary		Remove Resource from all Resource-Groups
//	@Description	Removes resource from all resource-groups it belongs to
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"User name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/remove-resource [post]
func RemoveResource(ctx *gin.Context) {
	// Get resource name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "GroupName cannot be empty"})

		return
	}

	err := auth.RemoveResource(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove resource from resource group")
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to remove resource from resource group"},
		)

		return
	}

	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Resource removed from all its Resource-Groups successfully!",
	})
}

// GetGroupsOfResource godoc
//
//	@Summary		Get Groups of a Resource
//	@Description	Returns all groups a resource belongs to
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"Resource name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/groups-of-resource [post]
func GetGroupsOfResource(ctx *gin.Context) {
	// Get group name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Resource name cannot be empty"})

		return
	}

	groups, err := auth.GetGroupsForResource(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to get groups of resource")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get groups of resource"})

		return
	}
	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"resource": name,
		"groups":   groups,
	})
}

// GetResourcesOfGroup godoc
//
//	@Summary		Get Resources of a Group
//	@Description	Returns all resources that belong to a specific group
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string				true	"Group name"
//	@Success		201		{object}	map[string]string	"TBD"
//	@Failure		400		{object}	map[string]string	"bad request"
//	@Failure		404		{object}	map[string]string	"not found"
//	@Failure		500		{object}	map[string]string	"internal server error"
//	@Router			/auth/resources-of-group [post]
func GetResourcesOfGroup(ctx *gin.Context) {
	// Get group name from path parameter
	name := ctx.Query("name")

	// Check that group name is not empty
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Group name cannot be empty"})

		return
	}
	resources, err := auth.GetResourceGroup(name)
	if err != nil {
		log.Error().Err(err).Msg("failed to get resources of group")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get resources of group"})

		return
	}
	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"group":     name,
		"resources": resources,
	})
}
