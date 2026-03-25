package api

import (
	"api-server/orm"
	"context"
	"errors"
	"strings"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/google/uuid"
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
			return strings.Compare(a.ID.String(), b.ID.String())
		},
	)

	usersParsed := make([]UserResponse, len(paginatedResult))
	for i, user := range paginatedResult {
		if (request.Params.Name == nil && request.Params.DisplayName == nil) ||
			(request.Params.Name != nil && *request.Params.Name == user.Username) ||
			(request.Params.DisplayName != nil && *request.Params.DisplayName == user.DisplayName) {

			roles, err := server.authModule.GetGroupsForUser(user.ID.String())
			if err != nil {
				log.Error().Err(err).Msg("Failed to get user roles")

				return GenericInternalServerErrorResponse{}, nil
			}
			usersParsed[i] = UserResponse{
				Id:          user.ID.String(),
				Name:        user.Username,
				DisplayName: user.DisplayName,
				Roles:       &roles,
			}
		}
	}

	return GetV1User200JSONResponse(usersParsed), nil
}

// GetV1UserId implements StrictServerInterface.
func (server *Server) GetV1UserId(ctx context.Context, request GetV1UserIdRequestObject) (GetV1UserIdResponseObject, error) {
	uuidParser, err := uuid.Parse(request.Id)
	if err != nil {
		return GetV1UserId400JSONResponse{
			GenericBadRequestJSONResponse{
				"Provided uuid is invalid",
			},
		}, nil
	}

	user, err := server.db.GetUserByID(ctx, uuidParser)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return GetV1UserId404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get user by ID")

		return GenericInternalServerErrorResponse{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.ID.String())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return GetV1UserId200JSONResponse{
		Id:          user.ID.String(),
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
	}, nil
}

// HeadV1UserId implements StrictServerInterface.
func (server *Server) HeadV1UserId(
	ctx context.Context,
	request HeadV1UserIdRequestObject,
) (HeadV1UserIdResponseObject, error) {
	uuidParser, err := uuid.Parse(request.Id)
	if err != nil {
		return HeadV1UserId400Response{}, nil
	}

	_, err = server.db.GetUserByID(ctx, uuidParser)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return HeadV1UserId404Response{}, nil
		}

		log.Error().Err(err).Msg("Failed to get user by ID")

		return HeadV1UserId500Response{}, nil
	}

	return HeadV1UserId200Response{}, nil
}

// PostV1User implements StrictServerInterface.
func (server *Server) PostV1User(
	ctx context.Context,
	request PostV1UserRequestObject,
) (PostV1UserResponseObject, error) {
	if strings.TrimSpace(request.Body.Name) == "" ||
		strings.TrimSpace(request.Body.Password) == "" ||
		strings.TrimSpace(request.Body.DisplayName) == "" {
		return PostV1User400JSONResponse{
			GenericBadRequestJSONResponse{
				"Username, password, and display name cannot be empty",
			},
		}, nil
	}

	user, err := server.db.CreateUser(
		ctx,
		request.Body.Name,
		request.Body.Password,
		request.Body.DisplayName,
	)
	if err != nil {
		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			return PostV1User409JSONResponse{
				errConflict.Conflict,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to create user")

		return GenericInternalServerErrorResponse{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.ID.String())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return PostV1User201JSONResponse{
		Id:          user.ID.String(),
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
	}, nil
}

// PatchV1UserId implements StrictServerInterface.
func (server *Server) PatchV1UserId(
	ctx context.Context,
	request PatchV1UserIdRequestObject,
) (PatchV1UserIdResponseObject, error) {
	uuidParser, err := uuid.Parse(request.Id)
	if err != nil {
		return PatchV1UserId400JSONResponse{
			GenericBadRequestJSONResponse{
				"Provided uuid is invalid",
			},
		}, nil
	}

	user, err := server.db.PatchUser(
		ctx,
		uuidParser,
		request.Body.Name,
		request.Body.Password,
		request.Body.DisplayName,
	)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return PatchV1UserId404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			return PatchV1UserId409JSONResponse{
				errConflict.Conflict,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to update user")

		return GenericInternalServerErrorResponse{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.ID.String())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return PatchV1UserId200JSONResponse(UserResponse{
		Id:          user.ID.String(),
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
	}), nil
}

func (server *Server) DeleteV1UserId(
	ctx context.Context,
	request DeleteV1UserIdRequestObject,
) (DeleteV1UserIdResponseObject, error) {
	uuidParser, err := uuid.Parse(request.Id)
	if err != nil {
		return DeleteV1UserId400JSONResponse{
			GenericBadRequestJSONResponse{
				"Provided uuid is invalid",
			},
		}, nil
	}

	user, err := server.db.DeleteUserByID(ctx, uuidParser)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return DeleteV1UserId404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to delete user")

		return GenericInternalServerErrorResponse{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.ID.String())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return DeleteV1UserId200JSONResponse(UserResponse{
		Id:          user.ID.String(),
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
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

	uuidParser, err := uuid.Parse(authenticatedUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID as UUID")

		return GenericInternalServerErrorResponse{}, nil
	}

	user, err := server.db.GetUserByID(ctx, uuidParser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by ID")

		return GenericInternalServerErrorResponse{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.ID.String())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return GetV1UserMe200JSONResponse(UserResponse{
		Id:          user.ID.String(),
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

	uuidParser, err := uuid.Parse(authenticatedUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID as UUID")

		return GenericInternalServerErrorResponse{}, nil
	}

	user, err := server.db.PatchUser(
		ctx,
		uuidParser,
		request.Body.Name,
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

	roles, err := server.authModule.GetGroupsForUser(user.ID.String())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return PatchV1UserMe200JSONResponse(UserResponse{
		Id:          user.ID.String(),
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

	uuidParser, err := uuid.Parse(authenticatedUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID as UUID")

		return GenericInternalServerErrorResponse{}, nil
	}

	user, err := server.db.DeleteUserByID(ctx, uuidParser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete user")

		return GenericInternalServerErrorResponse{}, nil
	}

	roles, err := server.authModule.GetGroupsForUser(user.ID.String())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user roles")

		return GenericInternalServerErrorResponse{}, nil
	}

	return DeleteV1UserMe200JSONResponse(UserResponse{
		Id:          user.ID.String(),
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Roles:       &roles,
	}), nil
}
