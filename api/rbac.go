package api

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/rs/zerolog/log"
)

const internalKeyword = "_INTERNAL"

// DeleteV1RbacPolicy implements StrictServerInterface.
func (server *Server) DeleteV1RbacPolicy(ctx context.Context, request DeleteV1RbacPolicyRequestObject) (DeleteV1RbacPolicyResponseObject, error) {
	err := server.authModule.RemovePolicy(
		request.Body.Role,
		request.Body.ResourceGroup,
		string(request.Body.Method),
	)
	if err != nil {
		if errors.Is(err, &auth.ConflictError{}) {
			return DeleteV1RbacPolicy409JSONResponse{
				Error: "Cannot delete enclave admin policy",
			}, nil
		}

		log.Error().Err(err).Msg("Failed to delete policy")

		return GenericInternalServerErrorResponse{}, nil
	}

	return DeleteV1RbacPolicy200Response{}, nil
}

// DeleteV1RbacResourceGroupResourceGroup implements StrictServerInterface.
func (server *Server) DeleteV1RbacResourceGroupResourceGroup(ctx context.Context, request DeleteV1RbacResourceGroupResourceGroupRequestObject) (DeleteV1RbacResourceGroupResourceGroupResponseObject, error) {
	resourceGroup, err := server.authModule.GetResourceGroup(request.ResourceGroup)
	if err != nil {
		if errors.Is(err, &auth.NotFoundError{}) {
			return DeleteV1RbacResourceGroupResourceGroup404JSONResponse{
				GenericNotFoundJSONResponse{"Provided resource group does not exist"},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get resource group")

		return GenericInternalServerErrorResponse{}, nil
	}
	err = server.authModule.RemoveResourceGroup(request.ResourceGroup)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete resource group")

		return GenericInternalServerErrorResponse{}, nil
	}

	return DeleteV1RbacResourceGroupResourceGroup200JSONResponse(ResourceGroupResource{
		Name:      request.ResourceGroup,
		Endpoints: resourceGroup,
	}), nil
}

// DeleteV1RbacRoleRole implements StrictServerInterface.
func (server *Server) DeleteV1RbacRoleRole(ctx context.Context, request DeleteV1RbacRoleRoleRequestObject) (DeleteV1RbacRoleRoleResponseObject, error) {
	role, err := server.authModule.GetUserGroup(request.Role)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			return DeleteV1RbacRoleRole404JSONResponse{
				GenericNotFoundJSONResponse{"Provided role does not exist"},
			}, nil
		}

		var errConflict *auth.ConflictError
		if errors.As(err, &errConflict) {
			return DeleteV1RbacRoleRole409JSONResponse{
				errConflict.Reason,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to delete role")

		return GenericInternalServerErrorResponse{}, nil
	}
	err = server.authModule.RemoveUserGroup(request.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete role")

		return GenericInternalServerErrorResponse{}, nil
	}

	return DeleteV1RbacRoleRole200JSONResponse(RoleResource{
		Name:  request.Role,
		Users: role,
	}), nil
}

// GetV1RbacPolicy implements StrictServerInterface.
func (server *Server) GetV1RbacPolicy(ctx context.Context, request GetV1RbacPolicyRequestObject) (GetV1RbacPolicyResponseObject, error) {
	policies, err := server.authModule.ListPolicies()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list policies")

		return GenericInternalServerErrorResponse{}, nil
	}

	// Filter out internal policies
	policies = slices.Collect(func(yield func(policy auth.Policy) bool) {
		for _, policy := range policies {
			if isSanitized(policy.ResourceGroup) {
				if !yield(policy) {
					return
				}
			}
		}
	})

	rolesPaginated := paginate(policies, *request.Params.Limit, *request.Params.Offset, func(a, b auth.Policy) int {
		if a.UserGroup != b.UserGroup {
			return cmp.Compare(a.UserGroup, b.UserGroup)
		}

		if a.ResourceGroup != b.ResourceGroup {
			return cmp.Compare(a.ResourceGroup, b.ResourceGroup)
		}

		return cmp.Compare(a.Permission, b.Permission)
	})

	rolesTransformed := make([]RBACPolicy, len(rolesPaginated))
	for i, role := range rolesPaginated {
		rolesTransformed[i] = RBACPolicy{
			Role:          role.UserGroup,
			ResourceGroup: role.ResourceGroup,
			Method:        RBACPolicyMethod(role.Permission),
		}
	}

	return GetV1RbacPolicy200JSONResponse(rolesTransformed), nil
}

// GetV1RbacResourceGroup implements StrictServerInterface.
func (server *Server) GetV1RbacResourceGroup(ctx context.Context, request GetV1RbacResourceGroupRequestObject) (GetV1RbacResourceGroupResponseObject, error) {
	resourceGroups, err := server.authModule.GetResourceGroups()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get resource groups")

		return GenericInternalServerErrorResponse{}, nil
	}

	resourceGroupsMap := make(map[string][]string)
	for _, rg := range resourceGroups {
		if !isSanitized(rg.GroupName) {
			continue
		}

		if resourceGroupsMap[rg.GroupName] == nil {
			resourceGroupsMap[rg.GroupName] = []string{}
		}
		resourceGroupsMap[rg.GroupName] = append(resourceGroupsMap[rg.GroupName], rg.ResourceName)
	}

	resourceGroupsTransformed := make([]ResourceGroupResource, 0, len(resourceGroupsMap))
	for groupName, endpoints := range resourceGroupsMap {
		resourceGroupsTransformed = append(resourceGroupsTransformed, ResourceGroupResource{
			Name:      groupName,
			Endpoints: endpoints,
		})
	}

	resourceGroupsPaginated := paginate(resourceGroupsTransformed, *request.Params.Limit, *request.Params.Offset, func(a, b ResourceGroupResource) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return GetV1RbacResourceGroup200JSONResponse(resourceGroupsPaginated), nil
}

// GetV1RbacResourceGroupResourceGroup implements StrictServerInterface.
func (server *Server) GetV1RbacResourceGroupResourceGroup(ctx context.Context, request GetV1RbacResourceGroupResourceGroupRequestObject) (GetV1RbacResourceGroupResourceGroupResponseObject, error) {
	endpoints, err := server.authModule.GetResourceGroup(request.ResourceGroup)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			return GetV1RbacResourceGroupResourceGroup404JSONResponse{
				GenericNotFoundJSONResponse{"Provided resource group does not exist"},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get resource group")

		return GenericInternalServerErrorResponse{}, nil
	}

	return GetV1RbacResourceGroupResourceGroup200JSONResponse(ResourceGroupResource{
		Name:      request.ResourceGroup,
		Endpoints: endpoints,
	}), nil
}

// GetV1RbacRole implements StrictServerInterface.
func (server *Server) GetV1RbacRole(ctx context.Context, request GetV1RbacRoleRequestObject) (GetV1RbacRoleResponseObject, error) {
	roles, err := server.authModule.GetUserGroups()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user groups")

		return GenericInternalServerErrorResponse{}, nil
	}

	rolesMap := make(map[string][]string)
	for _, role := range roles {
		if rolesMap[role.GroupName] == nil {
			rolesMap[role.GroupName] = []string{}
		}
		rolesMap[role.GroupName] = append(rolesMap[role.GroupName], role.UserName)
	}

	rolesTransformed := make([]RoleResource, 0, len(rolesMap))
	for roleName, users := range rolesMap {
		if !isSanitized(roleName) {
			continue
		}
		rolesTransformed = append(rolesTransformed, RoleResource{
			Name:  roleName,
			Users: users,
		})
	}

	rolesPaginated := paginate(rolesTransformed, *request.Params.Limit, *request.Params.Offset, func(a, b RoleResource) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return GetV1RbacRole200JSONResponse(rolesPaginated), nil
}

// GetV1RbacRoleRole implements StrictServerInterface.
func (server *Server) GetV1RbacRoleRole(ctx context.Context, request GetV1RbacRoleRoleRequestObject) (GetV1RbacRoleRoleResponseObject, error) {
	users, err := server.authModule.GetUserGroup(request.Role)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			return GetV1RbacRoleRole404JSONResponse{
				GenericNotFoundJSONResponse{"Provided role does not exist"},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get user group")

		return GenericInternalServerErrorResponse{}, nil
	}

	return GetV1RbacRoleRole200JSONResponse(RoleResource{
		Name:  request.Role,
		Users: users,
	}), nil
}

// HeadV1RbacResourceGroupResourceGroup implements StrictServerInterface.
func (server *Server) HeadV1RbacResourceGroupResourceGroup(ctx context.Context, request HeadV1RbacResourceGroupResourceGroupRequestObject) (HeadV1RbacResourceGroupResourceGroupResponseObject, error) {
	exists, err := server.authModule.ResourceGroupExists(request.ResourceGroup)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check if resource group exists")

		return HeadV1RbacResourceGroupResourceGroup500Response{}, nil
	}

	if !exists {
		return HeadV1RbacResourceGroupResourceGroup404Response{}, nil
	}

	return HeadV1RbacResourceGroupResourceGroup200Response{}, nil
}

// HeadV1RbacRoleRole implements StrictServerInterface.
func (server *Server) HeadV1RbacRoleRole(ctx context.Context, request HeadV1RbacRoleRoleRequestObject) (HeadV1RbacRoleRoleResponseObject, error) {
	exists, err := server.authModule.UserGroupExists(request.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check if role exists")

		return HeadV1RbacRoleRole500Response{}, nil
	}

	if !exists {
		return HeadV1RbacRoleRole404Response{}, nil
	}

	return HeadV1RbacRoleRole200Response{}, nil
}

// PutV1RbacPolicy implements StrictServerInterface.
func (server *Server) PutV1RbacPolicy(ctx context.Context, request PutV1RbacPolicyRequestObject) (PutV1RbacPolicyResponseObject, error) {
	if request.Body.ResourceGroup != string(Asterisk) {
		resourceGroupExists, err := server.authModule.ResourceGroupExists(request.Body.ResourceGroup)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check if resource group exists")

			return GenericInternalServerErrorResponse{}, nil
		}

		if !resourceGroupExists {
			return PutV1RbacPolicy404JSONResponse{
				FieldErrorJSONResponse{
					&[]ErrField{
						{
							Field: "resourceGroup",
							Error: "Resource Group does not exist",
						},
					},
				},
			}, nil
		}
	}

	if request.Body.Role != string(Asterisk) {
		roleExists, err := server.authModule.UserGroupExists(request.Body.Role)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check if role exists")

			return GenericInternalServerErrorResponse{}, nil
		}

		if !roleExists {
			return PutV1RbacPolicy404JSONResponse{
				FieldErrorJSONResponse{
					&[]ErrField{
						{
							Field: "role",
							Error: "Role does not exist",
						},
					},
				},
			}, nil
		}
	}

	err := server.authModule.AddPolicy(
		request.Body.Role,
		request.Body.ResourceGroup,
		string(request.Body.Method),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add policy")

		return GenericInternalServerErrorResponse{}, nil
	}

	return PutV1RbacPolicy201Response{}, nil
}

// PutV1RbacResourceGroupResourceGroup implements StrictServerInterface.
func (server *Server) PutV1RbacResourceGroupResourceGroup(ctx context.Context, request PutV1RbacResourceGroupResourceGroupRequestObject) (PutV1RbacResourceGroupResourceGroupResponseObject, error) {
	currentEndpoints, err := server.authModule.GetResourceGroup(request.ResourceGroup)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			err = server.authModule.CreateResourceGroup(request.ResourceGroup)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create resource group")

				return GenericInternalServerErrorResponse{}, nil
			}

			currentEndpoints = []string{}
		} else {
			log.Error().Err(err).Msg("Failed to get resource group")

			return GenericInternalServerErrorResponse{}, nil
		}
	}

	for _, endpoint := range currentEndpoints {
		if slices.Contains(request.Body.Endpoints, endpoint) {
			continue
		}

		err = server.authModule.RemoveResourceFromGroup(endpoint, request.ResourceGroup)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to remove endpoint %s from resource group %s", endpoint, request.ResourceGroup)

			return GenericInternalServerErrorResponse{}, nil
		}
	}

	for _, endpoint := range request.Body.Endpoints {
		if slices.Contains(currentEndpoints, endpoint) {
			continue
		}

		err = server.authModule.AddResourceToGroup(endpoint, request.ResourceGroup)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to add endpoint %s to resource group %s", endpoint, request.ResourceGroup)

			return GenericInternalServerErrorResponse{}, nil
		}
	}

	return PutV1RbacResourceGroupResourceGroup201JSONResponse(ResourceGroupResource{
		Name:      request.ResourceGroup,
		Endpoints: request.Body.Endpoints,
	}), nil
}

// PutV1RbacRoleRole implements StrictServerInterface.
func (server *Server) PutV1RbacRoleRole(ctx context.Context, request PutV1RbacRoleRoleRequestObject) (PutV1RbacRoleRoleResponseObject, error) {
	currentUsers, err := server.authModule.GetUserGroup(request.Role)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			err = server.authModule.CreateUserGroup(request.Role)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create role")

				return GenericInternalServerErrorResponse{}, nil
			}

			currentUsers = []string{}
		} else {
			log.Error().Err(err).Msg("Failed to get user group")

			return GenericInternalServerErrorResponse{}, nil
		}
	}

	for _, user := range currentUsers {
		if slices.Contains(request.Body.Users, user) {
			continue
		}

		err = server.authModule.RemoveUserFromGroup(user, request.Role)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to remove user %s from role %s", user, request.Role)

			return GenericInternalServerErrorResponse{}, nil
		}
	}

	for _, user := range request.Body.Users {
		if slices.Contains(currentUsers, user) {
			continue
		}

		err = server.authModule.AddUserToGroup(user, request.Role)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to add user %s to role %s", user, request.Role)

			return GenericInternalServerErrorResponse{}, nil
		}
	}

	return PutV1RbacRoleRole201JSONResponse(RoleResource{
		Name:  request.Role,
		Users: request.Body.Users,
	}), nil
}

// isSanitized checks if a group name should be visible to API consumers.
// Returns false for internal groups containing the "_INTERNAL" keyword.
func isSanitized(group string) bool {
	return !strings.Contains(group, internalKeyword)
}
