package main

import (
	"api-server/api"
	"api-server/config"
	"api-server/orm"
	proto_gen "api-server/proto_gen"
	"api-server/queue"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/EnclaveRunner/shareddeps"
	"github.com/EnclaveRunner/shareddeps/auth"
	shareddepsConfig "github.com/EnclaveRunner/shareddeps/config"
	"github.com/gin-gonic/gin"
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
		{Key: "artifact_registry.port", Value: 9876},

		{Key: "redis.host", Value: "redis"},
		//nolint:mnd // Default port of redis
		{Key: "redis.port", Value: 6379},
		{Key: "redis.db", Value: 0},

		//nolint:mnd // Arbitrary defaults for default pagination size
		{Key: "pagination.default", Value: 50},
		//nolint:mnd // Arbitrary defaults for maximum pagination size
		{Key: "pagination.maximum", Value: 100},

		//nolint:mnd // Default max retries for task
		{Key: "retry.max_retries", Value: 3},
		{Key: "retry.retention", Value: "24h"},
	}

	// load config and create server
	cfg := &config.AppConfig{}
	err := shareddepsConfig.PopulateAppConfig(
		cfg,
		"api-server",
		"v0.10.0",
		defaults...)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load app config")
	}

	ginServer := shareddeps.InitRESTServer(cfg)
	ginServer.Use(paginationValidationMiddleware(cfg.Pagination.Default, cfg.Pagination.Maximum))

	dbGorm := orm.InitDB(cfg)
	policyAdapter := orm.NewCasbinAdapter(dbGorm)
	authModule := auth.NewModule(policyAdapter)

	db := orm.NewDB(authModule, dbGorm)

	shareddeps.AddAuth(
		ginServer,
		authModule,
		shareddeps.Authentication{BasicAuthenticator: db.BasicAuthFunc()},
	)

	// Initialize admin user after auth system is ready
	db.InitAdminUser(cfg)

	// Initialize task queue
	queueClient := queue.NewQueueClient(cfg, &db)

	// Migrate RBAC policies, resource groups and roles
	MigrateRBAC(authModule)

	registryClient := proto_gen.NewRegistryServiceClient(
		shareddeps.InitGRPCClient(
			cfg.ArtifactRegistry.Host,
			cfg.ArtifactRegistry.Port,
		),
	)
	retentionDuration, err := time.ParseDuration(cfg.Retry.Retention)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to parse retention duration (invalid format)")
	}

	if cfg.Pagination.Default > cfg.Pagination.Maximum {
		log.Fatal().
			Msg("Default pagination size cannot be greater than maximum pagination size")
	}
	server := api.NewServer(
		authModule,
		db,
		cfg.Retry.MaxRetries,
		retentionDuration,
		queueClient,
		registryClient,
	)
	handler := api.NewStrictHandler(server, nil)
	api.RegisterHandlers(ginServer, handler)

	shareddeps.StartRESTServer(cfg, ginServer)
}

// Init needed and default RBAC policies, resource groups and roles
func MigrateRBAC(authModule auth.AuthModule) {
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
		err := authModule.CreateResourceGroup(group)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to create resource group: %s", group)
		}
	}

	// Create user groups
	for _, group := range userGroups {
		err := authModule.CreateUserGroup(group)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to create user group: %s", group)
		}
	}

	// Add resources to groups
	for _, mapping := range resourceMappings {
		err := authModule.AddResourceToGroup(mapping.Resource, mapping.Group)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Failed to add resource %s to group %s", mapping.Resource, mapping.Group)
		}
	}

	// Add policies
	for _, policy := range policies {
		err := authModule.AddPolicy(
			policy.UserGroup,
			policy.ResourceGroup,
			policy.Method,
		)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Failed to add policy: %s -> %s [%s]", policy.UserGroup, policy.ResourceGroup, policy.Method)
		}
	}
}

func paginationValidationMiddleware(defaultLimit, maxLimit int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limitStr := ctx.Query("limit")
		offsetStr := ctx.Query("offset")

		var limit, offset int
		if limitStr == "" {
			limit = defaultLimit
		} else {
			parsedLimit, err := strconv.Atoi(limitStr)
			if err != nil || parsedLimit < 1 || parsedLimit > maxLimit {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Invalid limit parameter. Must be an integer between 1 and %d", maxLimit),
				})
				return
			}
			limit = parsedLimit
		}

		if offsetStr == "" {
			offset = 0
		} else {
			parsedOffset, err := strconv.Atoi(offsetStr)
			if err != nil || parsedOffset < 0 {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Invalid offset parameter. Must be a non-negative integer.",
				})
				return
			}
			offset = parsedOffset
		}

		if limit > maxLimit {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Limit parameter cannot exceed %d", maxLimit),
			})
			return
		}

		// Insert validated pagination parameters back into the query for handlers to use
		params := ctx.Request.URL.Query()
		params.Set("limit", strconv.Itoa(limit))
		params.Set("offset", strconv.Itoa(offset))
		ctx.Request.URL.RawQuery = params.Encode()
	}
}
