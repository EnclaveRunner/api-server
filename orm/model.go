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

type VirtualTask struct {
	TaskID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"task_id"`
	CreatedOn     time.Time `gorm:"not null;autoCreateTime"                        json:"created_on"`
	LastUpdated   time.Time `gorm:"autoUpdateTime"                                 json:"last_updated"`
	LastAction    string    `gorm:"not null"                                       json:"last_action"`
	RunnerHost    string    `                                                      json:"runner_host"`
	Retries       int       `gorm:"default:0"                                      json:"retries"`
	MaxRetries    int       `gorm:"default:3"                                      json:"max_retries"`
	Retention     string    `gorm:"default:'24h'"                                  json:"retention"`
	Status        string    `gorm:"not null"                                       json:"status"`
	ResultPayload []byte    `                                                      json:"result_payload"`
}

// TableName specifies the table name for VirtualTask
func (VirtualTask) TableName() string {
	return "virtual_tasks"
}

type TaskLog struct {
	ID        uuid.UUID   `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	TaskID    uuid.UUID   `gorm:"type:uuid;not null"                             json:"task_id"`
	Task      VirtualTask `gorm:"constraint:OnDelete:CASCADE"`
	Timestamp time.Time   `gorm:"not null;autoCreateTime"                        json:"timestamp"`
	Status    string      `gorm:"not null"                                       json:"status"`
	Issuer    string      `                                                      json:"issuer"`
	Payload   []byte      `                                                      json:"payload"`
}

// TableName specifies the table name for TaskLog
func (TaskLog) TableName() string {
	return "task_logs"
}
