package api

import (
	"context"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/rs/zerolog/log"
)

// GetRbacEndpoint implements StrictServerInterface.
func (s *Server) GetRbacEndpoint(
	ctx context.Context,
	request GetRbacEndpointRequestObject,
) (GetRbacEndpointResponseObject, error) {
	groups, err := auth.GetGroupsForResource(request.Body.Endpoint)
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
	groupExists, err := auth.ResourceGroupExists(request.Body.ResourceGroupName)
	if err != nil {
		return nil, &EmptyInternalServerError{}
	}
	if !groupExists {
		return &PostRbacEndpoint404JSONResponse{
			GenericNotFoundJSONResponse{
				Error: "Provided resource group does not exist",
			},
		}, nil
	}

	err = auth.AddResourceToGroup(
		request.Body.Endpoint,
		request.Body.ResourceGroupName,
	)
	if err != nil {
		return nil, &EmptyInternalServerError{}
	}

	return &PostRbacEndpoint201Response{}, nil
}

// DeleteRbacEndpoint implements StrictServerInterface.
func (s *Server) DeleteRbacEndpoint(
	ctx context.Context,
	request DeleteRbacEndpointRequestObject,
) (DeleteRbacEndpointResponseObject, error) {
	groupExists, err := auth.ResourceGroupExists(*request.Body.ResourceGroup)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check if resource group exists")
		return nil, &EmptyInternalServerError{}
	}

	if !groupExists {
		return &DeleteRbacEndpoint404JSONResponse{
			GenericNotFoundJSONResponse{
				Error: "Provided resource group does not exist",
			},
		}, nil
	}

	err = auth.RemoveResourceFromGroup(
		request.Body.Endpoint,
		*request.Body.ResourceGroup,
	)
}

// GetRbacPolicy implements StrictServerInterface.
func (s *Server) GetRbacPolicy(
	ctx context.Context,
	request GetRbacPolicyRequestObject,
) (GetRbacPolicyResponseObject, error) {
	panic("unimplemented")
}

// PostRbacPolicy implements StrictServerInterface.
func (s *Server) PostRbacPolicy(
	ctx context.Context,
	request PostRbacPolicyRequestObject,
) (PostRbacPolicyResponseObject, error) {
	panic("unimplemented")
}

// DeleteRbacPolicy implements StrictServerInterface.
func (s *Server) DeleteRbacPolicy(
	ctx context.Context,
	request DeleteRbacPolicyRequestObject,
) (DeleteRbacPolicyResponseObject, error) {
	panic("unimplemented")
}

// GetRbacListResourceGroups implements StrictServerInterface.
func (s *Server) GetRbacListResourceGroups(
	ctx context.Context,
	request GetRbacListResourceGroupsRequestObject,
) (GetRbacListResourceGroupsResponseObject, error) {
	panic("unimplemented")
}

// GetRbacResourceGroup implements StrictServerInterface.
func (s *Server) GetRbacResourceGroup(
	ctx context.Context,
	request GetRbacResourceGroupRequestObject,
) (GetRbacResourceGroupResponseObject, error) {
	panic("unimplemented")
}

// HeadRbacResourceGroup implements StrictServerInterface.
func (s *Server) HeadRbacResourceGroup(
	ctx context.Context,
	request HeadRbacResourceGroupRequestObject,
) (HeadRbacResourceGroupResponseObject, error) {
	panic("unimplemented")
}

// PostRbacResourceGroup implements StrictServerInterface.
func (s *Server) PostRbacResourceGroup(
	ctx context.Context,
	request PostRbacResourceGroupRequestObject,
) (PostRbacResourceGroupResponseObject, error) {
	panic("unimplemented")
}

// DeleteRbacResourceGroup implements StrictServerInterface.
func (s *Server) DeleteRbacResourceGroup(
	ctx context.Context,
	request DeleteRbacResourceGroupRequestObject,
) (DeleteRbacResourceGroupResponseObject, error) {
	panic("unimplemented")
}

// GetRbacListRoles implements StrictServerInterface.
func (s *Server) GetRbacListRoles(
	ctx context.Context,
	request GetRbacListRolesRequestObject,
) (GetRbacListRolesResponseObject, error) {
	panic("unimplemented")
}

// GetRbacRole implements StrictServerInterface.
func (s *Server) GetRbacRole(
	ctx context.Context,
	request GetRbacRoleRequestObject,
) (GetRbacRoleResponseObject, error) {
	panic("unimplemented")
}

// HeadRbacRole implements StrictServerInterface.
func (s *Server) HeadRbacRole(
	ctx context.Context,
	request HeadRbacRoleRequestObject,
) (HeadRbacRoleResponseObject, error) {
	panic("unimplemented")
}

// PostRbacRole implements StrictServerInterface.
func (s *Server) PostRbacRole(
	ctx context.Context,
	request PostRbacRoleRequestObject,
) (PostRbacRoleResponseObject, error) {
	panic("unimplemented")
}

// DeleteRbacRole implements StrictServerInterface.
func (s *Server) DeleteRbacRole(
	ctx context.Context,
	request DeleteRbacRoleRequestObject,
) (DeleteRbacRoleResponseObject, error) {
	panic("unimplemented")
}

// GetRbacUser implements StrictServerInterface.
func (s *Server) GetRbacUser(
	ctx context.Context,
	request GetRbacUserRequestObject,
) (GetRbacUserResponseObject, error) {
	panic("unimplemented")
}

// PostRbacUser implements StrictServerInterface.
func (s *Server) PostRbacUser(
	ctx context.Context,
	request PostRbacUserRequestObject,
) (PostRbacUserResponseObject, error) {
	panic("unimplemented")
}

// DeleteRbacUser implements StrictServerInterface.
func (s *Server) DeleteRbacUser(
	ctx context.Context,
	request DeleteRbacUserRequestObject,
) (DeleteRbacUserResponseObject, error) {
	panic("unimplemented")
}
