package main

import (
	"api-server/config"
	"api-server/handlers"
	"api-server/orm"

	"github.com/EnclaveRunner/shareddeps"
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
	viper.SetDefault("databsae.username", "enclave")
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

	// health check to see if api-server is reachable / ready
	shareddeps.Server.GET("/ready", handlers.Ready)

	auth := shareddeps.Server.Group("/auth")

	// user-group management endpoints
	// create new user-group
	auth.POST("/create-ugroup", handlers.CreateUserGroup)
	// remove user-group
	auth.POST("/remove-ugroup", handlers.RemoveUserGroup)
	// get all user-groups
	auth.GET("/ugroups", handlers.GetUserGroups)
	// add a user to a user-group
	auth.POST("/add-to-ugroup", handlers.AddToUserGroup)
	// remove a user from a user-group
	auth.POST("/remove-from-ugroup", handlers.RemoveFromUserGroup)
	// removes a user entirely
	auth.POST("/remove-user", handlers.RemoveUser)
	// get all groups a user belongs to
	auth.POST("/groups-of", handlers.GetGroupsOfUser)
	// get all users of a group
	auth.POST("/users-of", handlers.GetUsersOfGroup)

	// resource-group management endpoints
	auth.POST("/create-rgroup", handlers.CreateResourceGroup)
	// remove resource-group
	auth.POST("/remove-rgroup", handlers.RemoveResourceGroup)
	// get all resource-groups
	auth.GET("/rgroups", handlers.GetResourceGroups)
	// add resource to specified resource-group
	auth.GET("/add-to-rgroup", handlers.AddToResourceGroup)

	orm.InitDB()

	shareddeps.Start()
}
