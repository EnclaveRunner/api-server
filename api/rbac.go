package api

import (
	"api-server/orm"
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const internalKeyword = "_INTERNAL"

// GetRbacEndpoint implements StrictServerInterface.
func (s *Server) GetRbacEndpoint(
	ctx context.Context,
	request GetRbacEndpointRequestObject,
) (GetRbacEndpointResponseObject, error) {
	groups, err := auth.GetGroupsForResource(request.Params.Endpoint)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get groups for resource")

		return nil, &EmptyInternalServerError{}
	}

	return (*GetRbacEndpoint200JSONResponse)(&groups), nil
}

// PostRbacEndpoint implements StrictServerInterface.
func (s *Server) PostRbacEndpoint(
	ctx context.Context,
	request PostRbacEndpointRequestObject,
) (PostRbacEndpointResponseObject, error) {
	groupExists, err := auth.ResourceGroupExists(request.Body.ResourceGroup)
	if err != nil {
		return nil, &EmptyInternalServerError{}
	}
	if !groupExists {
		return PostRbacEndpoint404JSONResponse{
			GenericNotFoundJSONResponse{
				Error: "Provided resource group does not exist",
			},
		}, nil
	}

	err = auth.AddResourceToGroup(
		request.Body.Endpoint,
		request.Body.ResourceGroup,
	)
	if err != nil {
		return nil, &EmptyInternalServerError{}
	}

	return PostRbacEndpoint201Response{}, nil
}

// DeleteRbacEndpoint implements StrictServerInterface.
func (s *Server) DeleteRbacEndpoint(
	ctx context.Context,
	request DeleteRbacEndpointRequestObject,
) (DeleteRbacEndpointResponseObject, error) {
	groupExists, err := auth.ResourceGroupExists(request.Body.ResourceGroup)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check if resource group exists")

		return nil, &EmptyInternalServerError{}
	}

	if !groupExists {
		return DeleteRbacEndpoint404JSONResponse{
			GenericNotFoundJSONResponse{
				Error: "Provided resource group does not exist",
			},
		}, nil
	}

	err = auth.RemoveResourceFromGroup(
		request.Body.Endpoint,
		request.Body.ResourceGroup,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete resources from group")

		return nil, &EmptyInternalServerError{}
	}

	return DeleteRbacEndpoint200Response{}, nil
}

// GetRbacPolicy implements StrictServerInterface.
func (s *Server) GetRbacPolicy(
	ctx context.Context,
	request GetRbacPolicyRequestObject,
) (GetRbacPolicyResponseObject, error) {
	policies, err := auth.ListPolicies()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list policies")

		return nil, &EmptyInternalServerError{}
	}

	policiesParsed := make([]RBACPolicy, 0, len(policies))
	for _, policy := range policies {
		if !isSanitized(policy.ResourceGroup) {
			continue
		}
		policiesParsed = append(policiesParsed, RBACPolicy{
			Role:          policy.UserGroup,
			ResourceGroup: policy.ResourceGroup,
			Permission:    RBACPolicyPermission(policy.Permission),
		})
	}

	return (*GetRbacPolicy200JSONResponse)(&policiesParsed), nil
}

// PostRbacPolicy implements StrictServerInterface.
func (s *Server) PostRbacPolicy(
	ctx context.Context,
	request PostRbacPolicyRequestObject,
) (PostRbacPolicyResponseObject, error) {
	var fieldErrors []ErrField

	if request.Body.ResourceGroup != string(Asterisk) {
		// Validate resourceGroup field
		resourceGroupExists, err := auth.ResourceGroupExists(
			request.Body.ResourceGroup,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check if resource group exists")

			return nil, &EmptyInternalServerError{}
		}

		if !resourceGroupExists {
			fieldErrors = append(
				fieldErrors,
				ErrField{
					Field: "resourceGroup",
					Error: "Resource Group does not exist",
				},
			)
		}
	}

	if request.Body.Role != string(Asterisk) {
		// Validate role field
		roleExists, err := auth.UserGroupExists(request.Body.Role)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check if role exists")

			return nil, &EmptyInternalServerError{}
		}

		if !roleExists {
			fieldErrors = append(
				fieldErrors,
				ErrField{Field: "role", Error: "Role does not exist"},
			)
		}
	}

	if len(fieldErrors) > 0 {
		return PostRbacPolicy404JSONResponse{
			FieldErrorJSONResponse{&fieldErrors},
		}, nil
	}

	err := auth.AddPolicy(
		request.Body.Role,
		request.Body.ResourceGroup,
		string(request.Body.Permission),
	)
	if err != nil {
		log.Error().Err(err).Msg("Adding policy failed")

		return nil, &EmptyInternalServerError{}
	}

	return PostRbacPolicy201Response{}, nil
}

// DeleteRbacPolicy implements StrictServerInterface.
func (s *Server) DeleteRbacPolicy(
	ctx context.Context,
	request DeleteRbacPolicyRequestObject,
) (DeleteRbacPolicyResponseObject, error) {
	err := auth.RemovePolicy(
		request.Body.Role,
		request.Body.ResourceGroup,
		string(request.Body.Permission),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove policy")

		return nil, &EmptyInternalServerError{}
	}

	return DeleteRbacPolicy200Response{}, nil
}

// GetRbacListResourceGroups implements StrictServerInterface.
func (s *Server) GetRbacListResourceGroups(
	ctx context.Context,
	request GetRbacListResourceGroupsRequestObject,
) (GetRbacListResourceGroupsResponseObject, error) {
	resourceGroups, err := auth.GetResourceGroups()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get resource groups")

		return nil, &EmptyInternalServerError{}
	}

	resourceGroupsParsed := []string{}
	for _, rg := range resourceGroups {
		if !slices.Contains(resourceGroupsParsed, rg.GroupName) &&
			isSanitized(rg.GroupName) {
			resourceGroupsParsed = append(resourceGroupsParsed, rg.GroupName)
		}
	}

	return (*GetRbacListResourceGroups200JSONResponse)(
		&resourceGroupsParsed,
	), nil
}

// GetRbacResourceGroup implements StrictServerInterface.
func (s *Server) GetRbacResourceGroup(
	ctx context.Context,
	request GetRbacResourceGroupRequestObject,
) (GetRbacResourceGroupResponseObject, error) {
	resources, err := auth.GetResourceGroup(request.Params.ResourceGroup)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			return GetRbacResourceGroup404JSONResponse{
				GenericNotFoundJSONResponse{"Provided resource group does not exist"},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get resource group")

		return nil, &EmptyInternalServerError{}
	}

	return (*GetRbacResourceGroup200JSONResponse)(&resources), nil
}

// HeadRbacResourceGroup implements StrictServerInterface.
func (s *Server) HeadRbacResourceGroup(
	ctx context.Context,
	request HeadRbacResourceGroupRequestObject,
) (HeadRbacResourceGroupResponseObject, error) {
	exists, err := auth.ResourceGroupExists(request.Body.ResourceGroup)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get resource group")

		return nil, &EmptyInternalServerError{}
	}

	if !exists {
		return HeadRbacResourceGroup404Response{}, nil
	}

	return HeadRbacResourceGroup200Response{}, nil
}

// PostRbacResourceGroup implements StrictServerInterface.
func (s *Server) PostRbacResourceGroup(
	ctx context.Context,
	request PostRbacResourceGroupRequestObject,
) (PostRbacResourceGroupResponseObject, error) {
	err := auth.CreateResourceGroup(request.Body.ResourceGroup)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create resource group")

		return nil, &EmptyInternalServerError{}
	}

	return PostRbacResourceGroup201Response{}, nil
}

// DeleteRbacResourceGroup implements StrictServerInterface.
func (s *Server) DeleteRbacResourceGroup(
	ctx context.Context,
	request DeleteRbacResourceGroupRequestObject,
) (DeleteRbacResourceGroupResponseObject, error) {
	err := auth.RemoveResourceGroup(request.Body.ResourceGroup)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			return DeleteRbacResourceGroup404JSONResponse{
				GenericNotFoundJSONResponse{"Provided resource group does not exist"},
			}, nil
		}

		return nil, &EmptyInternalServerError{}
	}

	return DeleteRbacResourceGroup200Response{}, nil
}

// GetRbacListRoles implements StrictServerInterface.
func (s *Server) GetRbacListRoles(
	ctx context.Context,
	request GetRbacListRolesRequestObject,
) (GetRbacListRolesResponseObject, error) {
	groups, err := auth.GetUserGroups()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user groups")

		return nil, &EmptyInternalServerError{}
	}

	roles := []string{}
	for _, ug := range groups {
		if !slices.Contains(roles, ug.GroupName) {
			roles = append(roles, ug.GroupName)
		}
	}

	return (*GetRbacListRoles200JSONResponse)(&roles), nil
}

// GetRbacRole implements StrictServerInterface.
func (s *Server) GetRbacRole(
	ctx context.Context,
	request GetRbacRoleRequestObject,
) (GetRbacRoleResponseObject, error) {
	users, err := auth.GetUserGroup(request.Params.Role)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			return GetRbacRole404JSONResponse{
				GenericNotFoundJSONResponse{"Provided role does not exist"},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get user groups")

		return nil, &EmptyInternalServerError{}
	}

	return (*GetRbacRole200JSONResponse)(&users), nil
}

// HeadRbacRole implements StrictServerInterface.
func (s *Server) HeadRbacRole(
	ctx context.Context,
	request HeadRbacRoleRequestObject,
) (HeadRbacRoleResponseObject, error) {
	roleExists, err := auth.UserGroupExists(request.Body.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check if role exists")

		return nil, &EmptyInternalServerError{}
	}

	if !roleExists {
		return HeadRbacRole404Response{}, nil
	}

	return HeadRbacRole200Response{}, nil
}

// PostRbacRole implements StrictServerInterface.
func (s *Server) PostRbacRole(
	ctx context.Context,
	request PostRbacRoleRequestObject,
) (PostRbacRoleResponseObject, error) {
	err := auth.CreateUserGroup(request.Body.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create role")

		return nil, &EmptyInternalServerError{}
	}

	return PostRbacRole201Response{}, nil
}

// DeleteRbacRole implements StrictServerInterface.
func (s *Server) DeleteRbacRole(
	ctx context.Context,
	request DeleteRbacRoleRequestObject,
) (DeleteRbacRoleResponseObject, error) {
	err := auth.RemoveUserGroup(request.Body.Role)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			return DeleteRbacRole404JSONResponse{
				GenericNotFoundJSONResponse{"Provided role does not exist"},
			}, nil
		}

		var errConflict *auth.ConflictError
		if errors.As(err, &errConflict) {
			return DeleteRbacRole409JSONResponse{
				errConflict.Reason,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to delete role")

		return nil, &EmptyInternalServerError{}
	}

	return DeleteRbacRole200Response{}, nil
}

// GetRbacUser implements StrictServerInterface.
func (s *Server) GetRbacUser(
	ctx context.Context,
	request GetRbacUserRequestObject,
) (GetRbacUserResponseObject, error) {
	uuidParsed, err := uuid.Parse(request.Params.UserId)
	if err != nil {
		return GetRbacUser400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid UUID format for userId",
			},
		}, nil
	}

	_, err = orm.GetUserByID(ctx, uuidParsed)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return GetRbacUser404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "User not found",
				},
			}, nil
		}

		return nil, &EmptyInternalServerError{}
	}

	users, err := auth.GetGroupsForUser(request.Params.UserId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get roles for user")

		return nil, &EmptyInternalServerError{}
	}

	return (*GetRbacUser200JSONResponse)(&users), nil
}

// PostRbacUser implements StrictServerInterface.
func (s *Server) PostRbacUser(
	ctx context.Context,
	request PostRbacUserRequestObject,
) (PostRbacUserResponseObject, error) {
	uuidParsed, err := uuid.Parse(request.Body.UserId)
	if err != nil {
		return PostRbacUser400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid UUID format for userId",
			},
		}, nil
	}

	_, err = orm.GetUserByID(ctx, uuidParsed)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return PostRbacUser404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "User not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get user by ID")

		return nil, &EmptyInternalServerError{}
	}

	err = auth.AddUserToGroup(request.Body.UserId, request.Body.Role)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			return PostRbacUser404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "User not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to add user to group")

		return nil, &EmptyInternalServerError{}
	}

	return PostRbacUser201Response{}, nil
}

// DeleteRbacUser implements StrictServerInterface.
func (s *Server) DeleteRbacUser(
	ctx context.Context,
	request DeleteRbacUserRequestObject,
) (DeleteRbacUserResponseObject, error) {
	uuidParsed, err := uuid.Parse(request.Body.UserId)
	if err != nil {
		return DeleteRbacUser400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid UUID format for userId",
			},
		}, nil
	}

	_, err = orm.GetUserByID(ctx, uuidParsed)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return DeleteRbacUser404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Specified user not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get user by ID")

		return nil, &EmptyInternalServerError{}
	}

	err = auth.RemoveUserFromGroup(request.Body.UserId, request.Body.Role)
	if err != nil {
		var errNotFound *auth.NotFoundError
		if errors.As(err, &errNotFound) {
			return DeleteRbacUser404JSONResponse{
				GenericNotFoundJSONResponse{"Specified role does not exist"},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to remove user from role")

		return nil, &EmptyInternalServerError{}
	}

	return DeleteRbacUser200Response{}, nil
}

func isSanitized(group string) bool {
	return !strings.Contains(group, internalKeyword)
}
