package main

import (
	"api-server/config"
	"github.com/EnclaveRunner/shareddeps"
)

// @title			Enclave API Server
// @version			v0.0.0
// @description	API server for the Enclave project
// @termsOfService	http://swagger.io/terms/
// @contact.name	API Support
// @license.name	GNU General Public License v3.0
// @license.url	https://www.gnu.org/licenses/gpl-3.0.html
// @host			localhost:8080
// @BasePath		/api/v1
func main() {
	shareddeps.Init(config.Cfg, "api-server", "v0.0.0", nil, nil)
}
