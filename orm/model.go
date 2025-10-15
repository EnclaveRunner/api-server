package orm

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username string    `gorm:"unique;not null"`
}

type Auth_Basic struct {
	gorm.Model

	User     User      `gorm:"constraint:OnDelete:CASCADE"`
	UserID   uuid.UUID `gorm:"type:uuid;not null"`
	Password []byte    `gorm:"not null"`
}
