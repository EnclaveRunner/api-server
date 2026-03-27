package api

import (
	"api-server/orm"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/rs/zerolog/log"
)

// GetV1User implements StrictServerInterface.
func (server *Server) GetV1User(
	ctx context.Context,
	request GetV1UserRequestObject,
) (GetV1UserResponseObject, error) {
	users, err := server.db.ListAllUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users")

		return GenericInternalServerErrorResponse{}, nil
	}

	paginatedResult := paginate(
		users,
		*request.Params.Limit,
		*request.Params.Offset,
		func(a, b orm.User) int {
			return strings.Compare(a.Username, b.Username)
		},
	)

	usersParsed := make([]UserResponse, len(paginatedResult))
	for i, user := range paginatedResult {
		if (request.Params.Name == nil && request.Params.DisplayName == nil) ||
			(request.Params.Name != nil && *request.Params.Name == user.Username) ||
			(request.Params.DisplayName != nil && *request.Params.DisplayName == user.DisplayName) {

			roles, err := server.authModule.GetGroupsForUser(user.Username)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get user roles")

				return GenericInternalServerErrorResponse{}, nil
			}
			usersParsed[i] = UserResponse{
				Name:        user.Username,
				DisplayName: user.DisplayName,
				Roles:       &roles,
			}
		}
	}

	filteredUsers := make([]UserResponse, 0, len(usersParsed))
	for _, user := range usersParsed {
		if user.Name != "" {
			filteredUsers = append(filteredUsers, user)
		}
	}

	return GetV1User200JSONResponse(filteredUsers), nil
}

// GetV1UserUsername implements StrictServerInterface.
func (server *Server) GetV1UserUsername(
	ctx context.Context,
	request GetV1UserUsernameRequestObject,
) (GetV1UserUsernameResponseObject, error) {
	user, err := server.db.GetUserByUsername(ctx, request.Username)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return GetV1UserUsername404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get user by username")

		return GetV1UserUsername500Response{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.Username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GetV1UserUsername500Response{}, nil
	}

	return GetV1UserUsername200JSONResponse{
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
	}, nil
}

// HeadV1UserUsername implements StrictServerInterface.
func (server *Server) HeadV1UserUsername(
	ctx context.Context,
	request HeadV1UserUsernameRequestObject,
) (HeadV1UserUsernameResponseObject, error) {
	_, err := server.db.GetUserByUsername(ctx, request.Username)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return HeadV1UserUsername404Response{}, nil
		}

		log.Error().Err(err).Msg("Failed to get user by username")

		return HeadV1UserUsername500Response{}, nil
	}

	return HeadV1UserUsername200Response{}, nil
}

// PutV1UserUsername implements StrictServerInterface.
func (server *Server) PutV1UserUsername(
	ctx context.Context,
	request PutV1UserUsernameRequestObject,
) (PutV1UserUsernameResponseObject, error) {
	if strings.TrimSpace(request.Username) == "" ||
		strings.TrimSpace(request.Body.Password) == "" ||
		strings.TrimSpace(request.Body.DisplayName) == "" {
		return PutV1UserUsername400JSONResponse{
			GenericBadRequestJSONResponse{
				"Username, password, and display name cannot be empty",
			},
		}, nil
	}

	if request.Body.Roles != nil {
		for _, role := range *request.Body.Roles {
			exists, err := server.authModule.UserGroupExists(role)
			if err != nil {
				log.Error().Err(err).Msg("Checking wether role exists failed")

				return GenericInternalServerErrorResponse{}, nil
			}

			if !exists {
				return PutV1UserUsername400JSONResponse{
					GenericBadRequestJSONResponse{
						Error: fmt.Sprintf("Role %s does not exist", role),
					},
				}, nil
			}
		}
	}

	user, err := server.db.CreateUser(
		ctx,
		request.Username,
		request.Body.Password,
		request.Body.DisplayName,
	)
	if err != nil {
		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			return PutV1UserUsername409JSONResponse{
				errConflict.Conflict,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to create user")

		return GenericInternalServerErrorResponse{}, nil
	}

	if request.Body.Roles != nil {
		for _, role := range *request.Body.Roles {
			err := server.authModule.AddUserToGroup(request.Username, role)
			if err != nil {
				log.Error().Err(err).Msg("Failed to add user to group")
			}
		}
	}

	return PutV1UserUsername201JSONResponse{
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       request.Body.Roles,
	}, nil
}

// PatchV1UserUsername implements StrictServerInterface.
func (server *Server) PatchV1UserUsername(
	ctx context.Context,
	request PatchV1UserUsernameRequestObject,
) (PatchV1UserUsernameResponseObject, error) {
	if request.Body.Roles != nil {
		for _, role := range *request.Body.Roles {
			exists, err := server.authModule.UserGroupExists(role)
			if err != nil {
				log.Error().Err(err).Msg("Failed to check role existence")

				return GenericInternalServerErrorResponse{}, nil
			}

			if !exists {
				return PatchV1UserUsername400JSONResponse{
					GenericBadRequestJSONResponse{
						Error: fmt.Sprintf("Role %s does not exist", role),
					},
				}, nil
			}
		}
	}

	user, err := server.db.PatchUser(
		ctx,
		request.Username,
		request.Body.Password,
		request.Body.DisplayName,
	)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return PatchV1UserUsername404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			return PatchV1UserUsername409JSONResponse{
				errConflict.Conflict,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to update user")

		return PatchV1UserUsername500Response{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.Username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return PatchV1UserUsername500Response{}, nil
	}

	if request.Body.Roles != nil {
		for _, role := range *request.Body.Roles {
			if !slices.Contains(roles, role) {
				err := server.authModule.AddUserToGroup(request.Username, role)
				if err != nil {
					log.Error().Err(err).Msg("Failed to add user to group")

					return GenericInternalServerErrorResponse{}, nil
				}
			}
		}
		for _, role := range roles {
			if !slices.Contains(*request.Body.Roles, role) {
				err := server.authModule.RemoveUserFromGroup(request.Username, role)
				if err != nil {
					log.Error().Err(err).Msg("Failed to remove user from role")

					return GenericInternalServerErrorResponse{}, nil
				}
			}
		}

		roles = *request.Body.Roles
	}

	return PatchV1UserUsername200JSONResponse(UserResponse{
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
	}), nil
}

func (server *Server) DeleteV1UserUsername(
	ctx context.Context,
	request DeleteV1UserUsernameRequestObject,
) (DeleteV1UserUsernameResponseObject, error) {
	assignedRoles, err := server.authModule.GetGroupsForUser(request.Username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return DeleteV1UserUsername500Response{}, nil
	}

	user, err := server.db.DeleteUserByUsername(ctx, request.Username)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return DeleteV1UserUsername404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to delete user")

		return DeleteV1UserUsername500Response{}, nil
	}

	return DeleteV1UserUsername200JSONResponse(UserResponse{
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &assignedRoles,
	}), nil
}

// GetV1UserMe implements StrictServerInterface.
func (server *Server) GetV1UserMe(
	ctx context.Context,
	request GetV1UserMeRequestObject,
) (GetV1UserMeResponseObject, error) {
	authenticatedUser := auth.GetAuthenticatedUser(ctx)
	if authenticatedUser == auth.UnauthenticatedUser {
		log.Debug().
			Any("userContext", authenticatedUser).
			Msg("Unauthenticated user tried to access /users/me endpoint")

		return GetV1UserMe401Response{}, nil
	}

	user, err := server.db.GetUserByUsername(ctx, authenticatedUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by username")

		return GenericInternalServerErrorResponse{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.Username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return GetV1UserMe200JSONResponse(UserResponse{
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
	}), nil
}

// PatchUsersMe implements StrictServerInterface.
func (server *Server) PatchV1UserMe(
	ctx context.Context,
	request PatchV1UserMeRequestObject,
) (PatchV1UserMeResponseObject, error) {
	authenticatedUser := auth.GetAuthenticatedUser(ctx)
	if authenticatedUser == auth.UnauthenticatedUser {
		return PatchV1UserMe401Response{}, nil
	}

	user, err := server.db.PatchUser(
		ctx,
		authenticatedUser,
		request.Body.Password,
		request.Body.DisplayName,
	)
	if err != nil {
		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			return PatchV1UserMe409JSONResponse{
				errConflict.Conflict,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to update user")

		return GenericInternalServerErrorResponse{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.Username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return PatchV1UserMe200JSONResponse(UserResponse{
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
	}), nil
}

// DeleteV1UserMe implements StrictServerInterface.
func (server *Server) DeleteV1UserMe(
	ctx context.Context,
	request DeleteV1UserMeRequestObject,
) (DeleteV1UserMeResponseObject, error) {
	authenticatedUser := auth.GetAuthenticatedUser(ctx)

	if authenticatedUser == auth.UnauthenticatedUser {
		return DeleteV1UserMe401Response{}, nil
	}

	user, err := server.db.DeleteUserByUsername(ctx, authenticatedUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user")

		return GenericInternalServerErrorResponse{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.Username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return DeleteV1UserMe200JSONResponse(UserResponse{
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
	}), nil
}
