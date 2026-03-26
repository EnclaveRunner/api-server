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
		pwd, err := gorm.G[Auth_Basic](
			db.dbGorm,
		).Where(&Auth_Basic{Username: username}).
			First(ctx)
		if err != nil {
			return "", gorm.ErrRecordNotFound
		}

		// validate password
		err = bcrypt.CompareHashAndPassword(pwd.Password, []byte(password))
		if err != nil {
			return "", gorm.ErrRecordNotFound
		}

		return username, nil
	}
}
