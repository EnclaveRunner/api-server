package orm

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Username string    `gorm:"unique;not null"                                json:"username"`
}

type Auth_Basic struct {
	gorm.Model

	User     User      `gorm:"constraint:OnDelete:CASCADE"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Password []byte    `gorm:"not null"`
}
