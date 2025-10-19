package api

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen-config.yml openapi.yml

import (
	"context"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ StrictServerInterface = (*Server)(nil)

type Server struct{}

// DeleteRbacEndpoint implements StrictServerInterface.
func (s *Server) DeleteRbacEndpoint(
	ctx context.Context,
	request DeleteRbacEndpointRequestObject,
) (DeleteRbacEndpointResponseObject, error) {
	panic("unimplemented")
}

// DeleteRbacPolicy implements StrictServerInterface.
func (s *Server) DeleteRbacPolicy(
	ctx context.Context,
	request DeleteRbacPolicyRequestObject,
) (DeleteRbacPolicyResponseObject, error) {
	panic("unimplemented")
}

// DeleteRbacResourceGroup implements StrictServerInterface.
func (s *Server) DeleteRbacResourceGroup(
	ctx context.Context,
	request DeleteRbacResourceGroupRequestObject,
) (DeleteRbacResourceGroupResponseObject, error) {
	panic("unimplemented")
}

// DeleteRbacRole implements StrictServerInterface.
func (s *Server) DeleteRbacRole(
	ctx context.Context,
	request DeleteRbacRoleRequestObject,
) (DeleteRbacRoleResponseObject, error) {
	panic("unimplemented")
}

// DeleteRbacUser implements StrictServerInterface.
func (s *Server) DeleteRbacUser(
	ctx context.Context,
	request DeleteRbacUserRequestObject,
) (DeleteRbacUserResponseObject, error) {
	panic("unimplemented")
}

// DeleteUsersUser implements StrictServerInterface.
func (s *Server) DeleteUsersUser(
	ctx context.Context,
	request DeleteUsersUserRequestObject,
) (DeleteUsersUserResponseObject, error) {
	panic("unimplemented")
}

// GetRbacEndpoint implements StrictServerInterface.
func (s *Server) GetRbacEndpoint(
	ctx context.Context,
	request GetRbacEndpointRequestObject,
) (GetRbacEndpointResponseObject, error) {
	panic("unimplemented")
}

// GetRbacListResourceGroups implements StrictServerInterface.
func (s *Server) GetRbacListResourceGroups(
	ctx context.Context,
	request GetRbacListResourceGroupsRequestObject,
) (GetRbacListResourceGroupsResponseObject, error) {
	panic("unimplemented")
}

// GetRbacListRoles implements StrictServerInterface.
func (s *Server) GetRbacListRoles(
	ctx context.Context,
	request GetRbacListRolesRequestObject,
) (GetRbacListRolesResponseObject, error) {
	panic("unimplemented")
}

// GetRbacPolicy implements StrictServerInterface.
func (s *Server) GetRbacPolicy(
	ctx context.Context,
	request GetRbacPolicyRequestObject,
) (GetRbacPolicyResponseObject, error) {
	panic("unimplemented")
}

// GetRbacResourceGroup implements StrictServerInterface.
func (s *Server) GetRbacResourceGroup(
	ctx context.Context,
	request GetRbacResourceGroupRequestObject,
) (GetRbacResourceGroupResponseObject, error) {
	panic("unimplemented")
}

// GetRbacRole implements StrictServerInterface.
func (s *Server) GetRbacRole(
	ctx context.Context,
	request GetRbacRoleRequestObject,
) (GetRbacRoleResponseObject, error) {
	panic("unimplemented")
}

// GetRbacUser implements StrictServerInterface.
func (s *Server) GetRbacUser(
	ctx context.Context,
	request GetRbacUserRequestObject,
) (GetRbacUserResponseObject, error) {
	panic("unimplemented")
}

// GetUsersList implements StrictServerInterface.
func (s *Server) GetUsersList(
	ctx context.Context,
	request GetUsersListRequestObject,
) (GetUsersListResponseObject, error) {
	panic("unimplemented")
}

// GetUsersMe implements StrictServerInterface.
func (s *Server) GetUsersMe(
	ctx context.Context,
	request GetUsersMeRequestObject,
) (GetUsersMeResponseObject, error) {
	panic("unimplemented")
}

// GetUsersUser implements StrictServerInterface.
func (s *Server) GetUsersUser(
	ctx context.Context,
	request GetUsersUserRequestObject,
) (GetUsersUserResponseObject, error) {
	panic("unimplemented")
}

// HeadRbacResourceGroup implements StrictServerInterface.
func (s *Server) HeadRbacResourceGroup(
	ctx context.Context,
	request HeadRbacResourceGroupRequestObject,
) (HeadRbacResourceGroupResponseObject, error) {
	panic("unimplemented")
}

// HeadRbacRole implements StrictServerInterface.
func (s *Server) HeadRbacRole(
	ctx context.Context,
	request HeadRbacRoleRequestObject,
) (HeadRbacRoleResponseObject, error) {
	panic("unimplemented")
}

// HeadUsersUser implements StrictServerInterface.
func (s *Server) HeadUsersUser(
	ctx context.Context,
	request HeadUsersUserRequestObject,
) (HeadUsersUserResponseObject, error) {
	panic("unimplemented")
}

// PatchUsersMe implements StrictServerInterface.
func (s *Server) PatchUsersMe(
	ctx context.Context,
	request PatchUsersMeRequestObject,
) (PatchUsersMeResponseObject, error) {
	panic("unimplemented")
}

// PatchUsersUser implements StrictServerInterface.
func (s *Server) PatchUsersUser(
	ctx context.Context,
	request PatchUsersUserRequestObject,
) (PatchUsersUserResponseObject, error) {
	panic("unimplemented")
}

// PostRbacEndpoint implements StrictServerInterface.
func (s *Server) PostRbacEndpoint(
	ctx context.Context,
	request PostRbacEndpointRequestObject,
) (PostRbacEndpointResponseObject, error) {
	panic("unimplemented")
}

// PostRbacPolicy implements StrictServerInterface.
func (s *Server) PostRbacPolicy(
	ctx context.Context,
	request PostRbacPolicyRequestObject,
) (PostRbacPolicyResponseObject, error) {
	panic("unimplemented")
}

// PostRbacResourceGroup implements StrictServerInterface.
func (s *Server) PostRbacResourceGroup(
	ctx context.Context,
	request PostRbacResourceGroupRequestObject,
) (PostRbacResourceGroupResponseObject, error) {
	panic("unimplemented")
}

// PostRbacRole implements StrictServerInterface.
func (s *Server) PostRbacRole(
	ctx context.Context,
	request PostRbacRoleRequestObject,
) (PostRbacRoleResponseObject, error) {
	panic("unimplemented")
}

// PostRbacUser implements StrictServerInterface.
func (s *Server) PostRbacUser(
	ctx context.Context,
	request PostRbacUserRequestObject,
) (PostRbacUserResponseObject, error) {
	panic("unimplemented")
}

// PostUsersUser implements StrictServerInterface.
func (s *Server) PostUsersUser(
	ctx context.Context,
	request PostUsersUserRequestObject,
) (PostUsersUserResponseObject, error) {
	panic("unimplemented")
}
