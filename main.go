package main

import (
	"api-server/api"
	"api-server/config"
	"api-server/handlers"
	"api-server/orm"

	"github.com/EnclaveRunner/shareddeps"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// @title			Enclave API Server
// @version			v0.0.0
// @description	API Central Entrypoint for the Enclave Platform
// @license.name	GNU General Public License v3.0
// @license.url	https://www.gnu.org/licenses/gpl-3.0.html
// @host			localhost:8080
func main() {
	// Set configuration defaults
	//nolint:mnd // Default port for PostgreSQL database
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.username", "enclave")
	viper.SetDefault("database.password", "enclave")
	viper.SetDefault("database.database", "enclave")

	// default credentials for admin / initial user
	viper.SetDefault("admin.username", "enclave")
	viper.SetDefault("admin.password", "enclave")

	// load config and create server
	shareddeps.Init(config.Cfg, "api-server", "v0.0.0")

	policyAdapter := orm.InitDB()

	shareddeps.AddAuth(
		policyAdapter,
		shareddeps.Authentication{BasicAuthenticator: orm.BasicAuth},
	)

	// Initialize admin user after auth system is ready
	orm.InitAdminUser()

	server := api.NewServer()
	api.RegisterHandlers(shareddeps.Server, server)

	// health check to see if api-server is reachable / ready
	shareddeps.Server.GET("/ready", handlers.Ready)

	shareddeps.Server.GET("/user", handlers.GetUser)
	shareddeps.Server.POST("/user", handlers.CreateUser)
	shareddeps.Server.PATCH("/user", handlers.PatchUser)
	shareddeps.Server.DELETE("/user", handlers.DeleteUser)

	shareddeps.Server.GET("/list-users", handlers.ListUsers)

	shareddeps.Server.GET("/me", handlers.GetMe)
	shareddeps.Server.PATCH("/me", handlers.UpdateMe)
	shareddeps.Server.DELETE("/me", handlers.DeleteMe)

	auth := shareddeps.Server.Group("/auth")

	// user-group management endpoints
	auth.POST("/ugroup", handlers.CreateUserGroup)
	auth.DELETE("/ugroup", handlers.RemoveUserGroup)
	auth.GET("/list-ugroups", handlers.GetUserGroups)
	auth.POST("/user", handlers.AddToUserGroup)
	auth.DELETE("/user", handlers.RemoveFromUserGroup)
	auth.GET("/groups-of-user", handlers.GetGroupsOfUser)
	auth.GET("/users-of-group", handlers.GetUsersOfGroup)

	// resource-group management endpoints
	auth.POST("/rgroup", handlers.CreateResourceGroup)
	auth.DELETE("/rgroup", handlers.RemoveResourceGroup)
	auth.GET("/list-rgroups", handlers.GetResourceGroups)
	auth.POST("/resource", handlers.AddToResourceGroup)
	auth.DELETE("/resource", handlers.RemoveFromResourceGroup)
	auth.GET("/groups-of-resource", handlers.GetGroupsOfResource)
	auth.GET("/resources-of-group", handlers.GetResourcesOfGroup)

	// policy management endpoints
	// create new policy
	auth.POST("/create-policy", handlers.CreatePolicy)
	// delete a policy
	auth.POST("/remove-policy", handlers.RemovePolicy)

	shareddeps.Start()
}
