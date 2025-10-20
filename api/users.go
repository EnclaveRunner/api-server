package api

import (
	"api-server/orm"
	"context"
	"errors"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// GetUsersList implements StrictServerInterface.
func (s *Server) GetUsersList(
	ctx context.Context,
	request GetUsersListRequestObject,
) (GetUsersListResponseObject, error) {
	users, err := orm.ListAllUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users")

		return nil, &EmptyInternalServerError{}
	}

	usersParsed := make([]UserResponse, len(users))
	for i, user := range users {
		usersParsed[i] = UserResponse{
			user.ID.String(),
			user.Username,
		}
	}

	return GetUsersList200JSONResponse(usersParsed), nil
}

// GetUsersUser implements StrictServerInterface.
func (s *Server) GetUsersUser(
	ctx context.Context,
	request GetUsersUserRequestObject,
) (GetUsersUserResponseObject, error) {
	uuidParser, err := uuid.Parse(request.Body.Id)
	if err != nil {
		return GetUsersUser400JSONResponse{
			GenericBadRequestJSONResponse{
				"Provided uuid is invalid",
			},
		}, nil
	}

	user, err := orm.GetUserByID(ctx, uuidParser)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return GetUsersUser404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get user by ID")

		return nil, &EmptyInternalServerError{}
	}

	return GetUsersUser200JSONResponse(UserResponse{
		user.ID.String(),
		user.Username,
	}), nil
}

// HeadUsersUser implements StrictServerInterface.
func (s *Server) HeadUsersUser(
	ctx context.Context,
	request HeadUsersUserRequestObject,
) (HeadUsersUserResponseObject, error) {
	uuidParser, err := uuid.Parse(request.Body.Id)
	if err != nil {
		return HeadUsersUser400JSONResponse{
			GenericBadRequestJSONResponse{
				"Provided uuid is invalid",
			},
		}, nil
	}

	_, err = orm.GetUserByID(ctx, uuidParser)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return HeadUsersUser404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get user by ID")

		return nil, &EmptyInternalServerError{}
	}

	return HeadUsersUser200Response{}, nil
}

// PostUsersUser implements StrictServerInterface.
func (s *Server) PostUsersUser(
	ctx context.Context,
	request PostUsersUserRequestObject,
) (PostUsersUserResponseObject, error) {
	user, err := orm.CreateUser(ctx, request.Body.Name, request.Body.Password)
	if err != nil {
		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			return PostUsersUser409JSONResponse{
				errConflict.Conflict,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to create user")

		return nil, &EmptyInternalServerError{}
	}

	return PostUsersUser201JSONResponse(UserResponse{
		user.ID.String(),
		user.Username,
	}), nil
}

// PatchUsersUser implements StrictServerInterface.
func (s *Server) PatchUsersUser(
	ctx context.Context,
	request PatchUsersUserRequestObject,
) (PatchUsersUserResponseObject, error) {
	uuidParser, err := uuid.Parse(request.Body.Id)
	if err != nil {
		return PatchUsersUser400JSONResponse{
			GenericBadRequestJSONResponse{
				"Provided uuid is invalid",
			},
		}, nil
	}

	user, err := orm.PatchUser(
		ctx,
		uuidParser,
		request.Body.NewName,
		request.Body.NewPassword,
	)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return PatchUsersUser404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			return PatchUsersUser409JSONResponse{
				errConflict.Conflict,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to update user")

		return nil, &EmptyInternalServerError{}
	}

	return PatchUsersUser200JSONResponse(UserResponse{
		user.ID.String(),
		user.Username,
	}), nil
}

func (s *Server) DeleteUsersUser(
	ctx context.Context,
	request DeleteUsersUserRequestObject,
) (DeleteUsersUserResponseObject, error) {
	uuidParser, err := uuid.Parse(request.Body.Id)
	if err != nil {
		return DeleteUsersUser400JSONResponse{
			GenericBadRequestJSONResponse{
				"Provided uuid is invalid",
			},
		}, nil
	}

	user, err := orm.DeleteUserByID(ctx, uuidParser)
	if err != nil {
		var errNotFound *orm.NotFoundError
		if errors.As(err, &errNotFound) {
			return DeleteUsersUser404JSONResponse{
				GenericNotFoundJSONResponse{
					"User not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to delete user")

		return nil, &EmptyInternalServerError{}
	}

	return DeleteUsersUser200JSONResponse(UserResponse{
		user.ID.String(),
		user.Username,
	}), nil
}

// GetUsersMe implements StrictServerInterface.
func (s *Server) GetUsersMe(
	ctx context.Context,
	request GetUsersMeRequestObject,
) (GetUsersMeResponseObject, error) {
	authenticatedUser := auth.RetrieveAuthenticatedUser(ctx)
	if authenticatedUser == auth.UnauthenticatedUser {
		log.Debug().Any("userContext", authenticatedUser).Msg("Unauthenticated user tried to access /users/me endpoint")

		log.Debug().Any("BasicAuthContext", ctx.Value(BasicAuthScopes)).Msg("Basic auth context for unauthenticated user")

		log.Debug().Any("ctx", ctx).Msg("context")
		return GetUsersMe401Response{}, nil
	}

	uuidParser, err := uuid.Parse(authenticatedUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID as UUID")

		return nil, &EmptyInternalServerError{}
	}

	user, err := orm.GetUserByID(ctx, uuidParser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by ID")

		return nil, &EmptyInternalServerError{}
	}

	return GetUsersMe200JSONResponse(UserResponse{
		user.ID.String(),
		user.Username,
	}), nil
}

// PatchUsersMe implements StrictServerInterface.
func (s *Server) PatchUsersMe(
	ctx context.Context,
	request PatchUsersMeRequestObject,
) (PatchUsersMeResponseObject, error) {
	authenticatedUser := auth.RetrieveAuthenticatedUser(ctx)
	if authenticatedUser == auth.UnauthenticatedUser {
		return PatchUsersMe401Response{}, nil
	}

	uuidParser, err := uuid.Parse(authenticatedUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse user ID as UUID")

		return nil, &EmptyInternalServerError{}
	}

	user, err := orm.PatchUser(
		ctx,
		uuidParser,
		request.Body.NewName,
		request.Body.NewPassword,
	)
	if err != nil {
		var errConflict *orm.ConflictError
		if errors.As(err, &errConflict) {
			return PatchUsersMe409JSONResponse{
				errConflict.Conflict,
			}, nil
		}

		log.Error().Err(err).Msg("Failed to update user")

		return nil, &EmptyInternalServerError{}
	}

	return PatchUsersMe200JSONResponse(UserResponse{
		user.ID.String(),
		user.Username,
	}), nil
}
