package orm

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Username string    `gorm:"unique;not null"                                json:"username"`
}

type Auth_Basic struct {
	User     User      `gorm:"constraint:OnDelete:CASCADE"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Password []byte    `gorm:"not null"`
}
