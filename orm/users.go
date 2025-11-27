package orm

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (db *DB) GetUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (*User, error) {
	user, err := gorm.G[User](db.dbGorm).Where(&User{ID: userID}).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{fmt.Sprintf("User with ID %s", userID)}
		} else {
			return nil, &DatabaseError{err}
		}
	}

	return &user, nil
}

func (db *DB) GetUserByUsername(
	ctx context.Context,
	username string,
) (*User, error) {
	user, err := gorm.G[User](db.dbGorm).Where(&User{Username: username}).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{"User with username " + username}
		} else {
			return nil, &DatabaseError{err}
		}
	}

	return &user, nil
}

func (db *DB) ListAllUsers(ctx context.Context) ([]User, error) {
	users, err := gorm.G[User](db.dbGorm).Find(ctx)
	if err != nil {
		return nil, &DatabaseError{err}
	}

	return users, nil
}

func (db *DB) CreateUser(
	ctx context.Context,
	username, password, displayName string,
) (*User, error) {
	user := User{
		Username:    username,
		DisplayName: displayName,
	}

	err := db.dbGorm.Transaction(func(tx *gorm.DB) error {
		_, err := gorm.G[User](tx).
			Where(&User{Username: username}).
			First(ctx)

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return &DatabaseError{err}
		}

		if err == nil {
			return &ConflictError{
				fmt.Sprintf("User with username %s already exists", username),
			}
		}

		err = gorm.G[User](tx).Create(ctx, &user)
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return &ConflictError{
					fmt.Sprintf("User with username %s already exists", username),
				}
			}

			return &DatabaseError{err}
		}

		log.Info().
			Str("username", user.Username).
			Str("id", user.ID.String()).
			Msg("Created new user")

		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(password),
			HashCost,
		)
		if err != nil {
			return &GenericError{err}
		}

		authRecord := Auth_Basic{
			UserID:   user.ID,
			Password: hashedPassword,
		}

		err = gorm.G[Auth_Basic](tx).Create(ctx, &authRecord)
		if err != nil {
			return &DatabaseError{err}
		}

		return nil
	})
	if err != nil {
		return nil, &GenericError{err}
	}

	return &user, nil
}

func (db *DB) PatchUser(
	ctx context.Context,
	userID uuid.UUID,
	newUsername, newPassword, newDisplayName *string,
) (*User, error) {
	user, err := db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = db.dbGorm.Transaction(func(tx *gorm.DB) error {
		updated := false

		if newUsername != nil {
			log.Info().
				Str("oldUsername", user.Username).
				Str("newUsername", *newUsername).
				Msg("Updating username")
			user.Username = *newUsername
			updated = true
		}

		if newDisplayName != nil {
			log.Info().
				Str("oldDisplayName", user.DisplayName).
				Str("newDisplayName", *newDisplayName).
				Msg("Updating display name")
			user.DisplayName = *newDisplayName
			updated = true
		}

		if updated {
			err := tx.Save(user).Error
			if err != nil {
				if errors.Is(err, gorm.ErrDuplicatedKey) {
					return &ConflictError{
						fmt.Sprintf("User with username %s already exists", *newUsername),
					}
				}

				return &DatabaseError{err}
			}

			log.Info().
				Str("username", user.Username).
				Str("displayName", user.DisplayName).
				Str("id", user.ID.String()).
				Msg("User updated successfully")
		}

		if newPassword != nil {
			hashedPassword, err := bcrypt.GenerateFromPassword(
				[]byte(*newPassword),
				HashCost,
			)
			if err != nil {
				return &GenericError{err}
			}

			authRecord := Auth_Basic{
				UserID:   user.ID,
				Password: hashedPassword,
			}

			err = tx.Save(authRecord).Error
			if err != nil {
				return &DatabaseError{err}
			}
		}

		return nil
	})
	if err != nil {
		return nil, &GenericError{err}
	}

	return user, nil
}

func (db *DB) DeleteUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (*User, error) {
	user, err := db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = db.authModule.RemoveUser(user.ID.String())
	if err != nil {
		return nil, &GenericError{err}
	}

	_, err = gorm.G[User](db.dbGorm).Where(user).Delete(ctx)
	if err != nil {
		return nil, &DatabaseError{err}
	}

	return user, nil
}
