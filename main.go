package main

import (
	"api-server/api"
	"api-server/config"
	"api-server/orm"
	proto_gen "api-server/proto_gen"
	"api-server/queue"

	"github.com/EnclaveRunner/shareddeps"
	"github.com/EnclaveRunner/shareddeps/auth"
	shareddepsConfig "github.com/EnclaveRunner/shareddeps/config"
	"github.com/rs/zerolog/log"
)

func main() {
	// Set configuration defaults
	defaults := []shareddepsConfig.DefaultValue{
		//nolint:mnd // Default port of postgres
		{Key: "database.port", Value: 5432},
		{Key: "database.host", Value: "postgres"},
		{Key: "database.sslmode", Value: "disable"},
		{Key: "database.username", Value: "enclave_user"},
		{Key: "database.password", Value: "enclave_password"},
		{Key: "database.database", Value: "enclave_db"},
		{Key: "admin.username", Value: "enclave"},
		{Key: "admin.password", Value: "enclave"},
		{Key: "admin.display_name", Value: "System Admin"},
		{Key: "artifact_registry.host", Value: "artifactregistry"},
		//nolint:mnd // Default port of artifact registry
		{Key: "artifact_registry.port", Value: 5000},

		{Key: "redis.host", Value: "redis"},
		//nolint:mnd // Default port of redis
		{Key: "redis.port", Value: 6379},
		{Key: "redis.db", Value: 0},
	}

	// load config and create server
	shareddeps.InitRESTServer(config.Cfg, "api-server", "v0.5.1", defaults...)

	policyAdapter := orm.InitDB()

	shareddeps.AddAuth(
		policyAdapter,
		shareddeps.Authentication{BasicAuthenticator: orm.BasicAuth},
	)

	// Initialize admin user after auth system is ready
	orm.InitAdminUser()

	// Initialize task queue
	queue.Init()

	// Migrate RBAC policies, resource groups and roles
	MigrateRBAC()

	server := api.NewServer()
	handler := api.NewStrictHandler(server, nil)
	api.RegisterHandlers(shareddeps.RESTServer, handler)

	shareddeps.InitGRPCClient(
		config.Cfg.ArtifactRegistry.Host,
		config.Cfg.ArtifactRegistry.Port,
	)

	proto_gen.Client = proto_gen.NewRegistryServiceClient(shareddeps.GRPCClient)

	shareddeps.StartRESTServer()
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
