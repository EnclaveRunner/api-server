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

	// load config and create server
	shareddeps.Init(config.Cfg, "api-server", "v0.0.0", nil, nil)

	// health check to see if api-server is reachable / ready
	shareddeps.Server.GET("/ready", handlers.Ready)

	orm.InitDB()

	shareddeps.Start()
	
}
