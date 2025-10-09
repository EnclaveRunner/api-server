package main

import (
	"api-server/config"
	"api-server/handlers"

	"github.com/EnclaveRunner/shareddeps"
)

// @title			Enclave API Server
// @version			v0.0.0
// @description	API Central Entrypoint for the Enclave Platform
// @license.name	GNU General Public License v3.0
// @license.url	https://www.gnu.org/licenses/gpl-3.0.html
// @host			localhost:8080
func main() {
	// load config and create server
	shareddeps.Init(config.Cfg, "api-server", "v0.0.0", nil, nil)

	// health check to see if api-server is reachable / ready
	shareddeps.Server.GET("/ready", handlers.Ready)

	shareddeps.Start()
}
