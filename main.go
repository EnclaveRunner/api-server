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

	// health check to see if api-server is reachable / ready
	shareddeps.Server.GET("/ready", handlers.Ready)

	shareddeps.Server.POST("/user", handlers.CreateUser)
	shareddeps.Server.DELETE("/user", handlers.DeleteUser)
	shareddeps.Server.PATCH("/user", handlers.UpdateUser)
	shareddeps.Server.GET("/user", handlers.GetUser)
	
	shareddeps.Server.GET("list-users", handlers.ListUsers)

	shareddeps.Server.DELETE("/me", handlers.DeleteMe)
	shareddeps.Server.PATCH("/me", handlers.UpdateMe)
	shareddeps.Server.GET("/me", handlers.GetMe)

	auth := shareddeps.Server.Group("/auth")

	// create a new user
	auth.POST("/create-user", handlers.CreateUser)
	// change password and / or username of a user
	shareddeps.Server.POST("/update-user", handlers.UpdateUser)
	// removes a user entirely
	auth.POST("/remove-user", handlers.RemoveUser)

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
	// get all groups a user belongs to
	auth.POST("/groups-of-user", handlers.GetGroupsOfUser)
	// get all users of a group
	auth.POST("/users-of-group", handlers.GetUsersOfGroup)

	// resource-group management endpoints
	auth.POST("/create-rgroup", handlers.CreateResourceGroup)
	// remove resource-group
	auth.POST("/remove-rgroup", handlers.RemoveResourceGroup)
	// get all resource-groups
	auth.GET("/rgroups", handlers.GetResourceGroups)
	// add resource to specified resource-group
	auth.POST("/add-to-rgroup", handlers.AddToResourceGroup)
	// remove a user from a user-group
	auth.POST("/remove-from-rgroup", handlers.RemoveFromResourceGroup)
	// removes a resource from all groups it belongs to
	auth.POST("/remove-resource", handlers.RemoveResource)
	// get all groups a resource belongs to
	auth.POST("/groups-of-resource", handlers.GetGroupsOfResource)
	// get resources of a group
	auth.POST("/resources-of-group", handlers.GetResourcesOfGroup)

	// policy management endpoints
	// create new policy
	auth.POST("/create-policy", handlers.CreatePolicy)
	// delete a policy
	auth.POST("/remove-policy", handlers.RemovePolicy)

	shareddeps.Start()
}
