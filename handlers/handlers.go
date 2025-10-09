package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Demo godoc
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
