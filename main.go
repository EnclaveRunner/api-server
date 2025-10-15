package main

import (
	"api-server/config"
	"api-server/handlers"
	"api-server/orm"
	"encoding/csv"
	"os"

	"github.com/EnclaveRunner/shareddeps"
	"github.com/rs/zerolog/log"
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

	defaultPolicies, err := loadDefaults("default-policies.csv")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load default policies")
	}

	defaultUserGroups, err := loadDefaults("default-user-group-definitions.csv")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load default user group definitions")
	}

	defaultRessourceGroups, err := loadDefaults("default-ressource-group-definitions.csv")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load default ressource group definitions")
	}

	policyAdapter := orm.InitDB()

	shareddeps.AddAuth(
		policyAdapter,
		defaultPolicies,
		defaultUserGroups,
		defaultRessourceGroups,
		shareddeps.Authentication{BasicAuthenticator: orm.BasicAuth},
	)

	// health check to see if api-server is reachable / ready
	shareddeps.Server.GET("/ready", handlers.Ready)

	orm.InitDB()

	shareddeps.Start()
}

func loadDefaults(fileName string) (records [][]string, err error) {
	// load default policies from csv file
	// Open the CSV file
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening " + fileName + " :")
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)

	// Read all records
	policies, err := reader.ReadAll()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read default policies from CSV")
	}

	return policies, nil
}
