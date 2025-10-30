package main

import (
	"api-server/api"
	"api-server/config"
	"api-server/orm"

	"github.com/EnclaveRunner/shareddeps"
	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	// Set configuration defaults
	//nolint:mnd // Default port for PostgreSQL database
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.username", "enclave_user")
	viper.SetDefault("database.password", "enclave_password")
	viper.SetDefault("database.database", "enclave_db")

	// default credentials for admin / initial user
	viper.SetDefault("admin.username", "enclave")
	viper.SetDefault("admin.password", "enclave")

	// load config and create server
	shareddeps.Init(config.Cfg, "api-server", "v0.4.0")

	policyAdapter := orm.InitDB()

	shareddeps.AddAuth(
		policyAdapter,
		shareddeps.Authentication{BasicAuthenticator: orm.BasicAuth},
	)

	// Initialize admin user after auth system is ready
	orm.InitAdminUser()

	// Migrate RBAC policies, resource groups and roles
	MigrateRBAC()

	server := api.NewServer()
	handler := api.NewStrictHandler(server, nil)
	api.RegisterHandlers(shareddeps.Server, handler)

	shareddeps.Start()
}

// Init needed and default RBAC policies, resource groups and roles
func MigrateRBAC() {
	resourceGroups := []string{
		"self_INTERNAL",
		"user_management",
		"rbac_management",
	}

	userGroups := []string{
		"read_only",
		"user_management",
		"rbac_management",
	}

	// Define resource to group mappings
	type ResourceMapping struct {
		Resource string
		Group    string
	}
	resourceMappings := []ResourceMapping{
		{"/users/me", "self_INTERNAL"},
		{"/users/user", "user_management"},
		{"/users/list", "user_management"},
		{"/rbac/list-roles", "rbac_management"},
		{"/rbac/role", "rbac_management"},
		{"/rbac/user", "rbac_management"},
		{"/rbac/list-resource-groups", "rbac_management"},
		{"/rbac/resource-group", "rbac_management"},
		{"/rbac/endpoint", "rbac_management"},
		{"/rbac/policy", "rbac_management"},
	}

	// Define policies
	type Policy struct {
		UserGroup     string
		ResourceGroup string
		Method        string
	}
	policies := []Policy{
		{"*", "self_INTERNAL", "*"},
		{"read_only", "*", "GET"},
		{"read_only", "*", "HEAD"},
		{"user_management", "user_management", "*"},
		{"rbac_management", "rbac_management", "*"},
	}

	// Create resource groups
	for _, group := range resourceGroups {
		err := auth.CreateResourceGroup(group)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to create resource group: %s", group)
		}
	}

	// Create user groups
	for _, group := range userGroups {
		err := auth.CreateUserGroup(group)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to create user group: %s", group)
		}
	}

	// Add resources to groups
	for _, mapping := range resourceMappings {
		err := auth.AddResourceToGroup(mapping.Resource, mapping.Group)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Failed to add resource %s to group %s", mapping.Resource, mapping.Group)
		}
	}

	// Add policies
	for _, policy := range policies {
		err := auth.AddPolicy(policy.UserGroup, policy.ResourceGroup, policy.Method)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Failed to add policy: %s -> %s [%s]", policy.UserGroup, policy.ResourceGroup, policy.Method)
		}
	}
}
