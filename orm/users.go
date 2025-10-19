package orm

import (
	"context"
	"errors"
	"fmt"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetUserByID(userID uuid.UUID) (*User, error) {
	user, err := gorm.G[User](DB).Where(&User{ID: userID}).
		First(context.Background())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &NotFoundError{fmt.Sprintf("User with ID %s", userID)}
		} else {
			return nil, &DatabaseError{err}
		}
	}

	return &user, nil
}

func ListAllUsers() ([]User, error) {
	users, err := gorm.G[User](DB).Find(context.Background())
	if err != nil {
		return nil, &DatabaseError{err}
	}

	return users, nil
}

func CreateUser(username, password string) (*User, error) {
	user := User{
		Username: username,
	}

	var createdUser User

	err := DB.Transaction(func(tx *gorm.DB) error {
		_, err := gorm.G[User](DB).
			Where(&User{Username: username}).
			First(context.Background())

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return &DatabaseError{err}
		}

		if err == nil {
			return &ConflictError{fmt.Sprintf("User with username %s already exists", username)}
		}

		err = gorm.G[User](DB).Create(context.Background(), &user)
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return &ConflictError{fmt.Sprintf("User with username %s already exists", username)}
			}

			return &DatabaseError{err}
		}

		createdUser, err = gorm.G[User](DB).
			Where(&User{Username: username}).
			First(context.Background())
		if err != nil {
			return &DatabaseError{err}
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), HashCost)
		if err != nil {
			return &GenericError{err}
		}

		authRecord := Auth_Basic{
			UserID:   createdUser.ID,
			Password: hashedPassword,
		}

		err = gorm.G[Auth_Basic](DB).Create(context.Background(), &authRecord)
		if err != nil {
			return &DatabaseError{err}
		}

		return nil
	})
	if err != nil {
		return nil, &GenericError{err}
	}

	return &createdUser, nil
}

func PatchUser(userID uuid.UUID, newUsername, newPassword *string) (*User, error) {
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		if newUsername != nil {
			user.Username = *newUsername
			err := tx.Save(user).Error
			if err != nil {
				if errors.Is(err, gorm.ErrDuplicatedKey) {
					return &ConflictError{fmt.Sprintf("User with username %s already exists", *newUsername)}
				}

				return &DatabaseError{err}
			}
		}

		if newPassword != nil {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*newPassword), HashCost)
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

func DeleteUserByID(userID uuid.UUID) (*User, error) {
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	err = auth.RemoveUser(user.ID.String())
	if err != nil {
		return nil, &GenericError{err}
	}

	_, err = gorm.G[User](DB).Where(user).Delete(context.Background())
	if err != nil {
		return nil, &DatabaseError{err}
	}

	return user, nil
}
