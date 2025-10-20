package orm

import (
	"context"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const HashCost = 10

func BasicAuth(ctx context.Context, username, password string) (string, error) {
	// find user in user table
	user, err := gorm.G[User](
		DB,
	).Where(&User{Username: username}).
		First(ctx)
	if err != nil {
		return "", gorm.ErrRecordNotFound
	}

	pwd, err := gorm.G[Auth_Basic](
		DB,
	).Where(&Auth_Basic{UserID: user.ID}).
		First(ctx)
	if err != nil {
		return "", gorm.ErrRecordNotFound
	}

	// validate password
	err = bcrypt.CompareHashAndPassword(pwd.Password, []byte(password))
	if err != nil {
		return "", gorm.ErrRecordNotFound
	}

	return user.ID.String(), nil
}
