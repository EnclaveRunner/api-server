package orm

import (
	"context"

	"github.com/EnclaveRunner/shareddeps/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const HashCost = 10

func (db *DB) BasicAuthFunc() middleware.BasicAuthenticator {
	return func(ctx context.Context, username, password string) (string, error) {
		// find user in user table
		user, err := gorm.G[User](
			db.dbGorm,
		).Where(&User{Username: username}).
			First(ctx)
		if err != nil {
			return "", gorm.ErrRecordNotFound
		}

		pwd, err := gorm.G[Auth_Basic](
			db.dbGorm,
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
}
