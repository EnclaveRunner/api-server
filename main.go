package main

import (
	"api-server/api"
	"api-server/config"
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

	server := api.NewServer()
	handler := api.NewStrictHandler(server, nil)
	api.RegisterHandlers(shareddeps.Server, handler)

	shareddeps.Start()
	<let ci terribly fail>
}
