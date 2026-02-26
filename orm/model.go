package orm

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Username    string    `gorm:"unique;not null"                                json:"username"`
	DisplayName string    `gorm:"not null"                                       json:"displayName"`
}

type Auth_Basic struct {
	User     User      `gorm:"constraint:OnDelete:CASCADE"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Password []byte    `gorm:"not null"`
}

type TaskLog struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	TaskID    string    `gorm:"not null"                                       json:"task_id"`
	Timestamp time.Time `gorm:"not null;autoCreateTime"                        json:"timestamp"`
	Level     string    `gorm:"not null"                                       json:"level"`
	Issuer    string    `gorm:"not null"                                       json:"issuer"`
	Message   string    `gorm:"not null"                                       json:"payload"`
}

// TableName specifies the table name for TaskLog
func (TaskLog) TableName() string {
	return "task_logs"
}
