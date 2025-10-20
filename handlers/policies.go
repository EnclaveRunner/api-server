package handlers

import (
	"net/http"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/gin-gonic/gin"

	"github.com/rs/zerolog/log"
)

// CreatePolicy godoc
//
// @Summary		Creates a new Policy
// @Description	Adds policy. Every parameter can be "*" as wildcard. userGroup
// and resourceGroup
// @Description	must exist.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			CreatePolicyBody	body		CreatePolicyBody	true	"body"
// @Success		201		{object}	map[string]string	"TBD"
// @Failure		400		{object}	map[string]string	"bad request"
// @Failure		404		{object}	map[string]string	"not found"
// @Failure		500		{object}	map[string]string	"internal server error"
// @Router			/auth/create-policy [post]
func CreatePolicy(ctx *gin.Context) {
	var body CreatePolicyBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	err := auth.AddPolicy(body.UserGroup, body.ResourceGroup, body.Action)
	if err != nil {
		log.Error().Err(err).Msg("failed to create policy")
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to create policy"},
		)

		return
	}
	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Policy created successfully!",
	})
}

// RemovePolicy godoc
//
// @Summary		Removes a Policy
// @Description	Removes policy. Every parameter can be "*" as wildcard.
// @Description	userGroup and resourceGroup must exist.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			CreatePolicyBody	body		CreatePolicyBody	true	"body"
// @Success		201		{object}	map[string]string	"TBD"
// @Failure		400		{object}	map[string]string	"bad request"
// @Failure		404		{object}	map[string]string	"not found"
// @Failure		500		{object}	map[string]string	"internal server error"
// @Router			/auth/remove-policy [post]
func RemovePolicy(ctx *gin.Context) {
	var body CreatePolicyBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	err := auth.RemovePolicy(body.UserGroup, body.ResourceGroup, body.Action)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove policy")
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to remove policy"},
		)

		return
	}
	// Success response
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Policy removed successfully!",
	})
}
