package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Project struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title       string         `gorm:"uniqueIndex" json:"title"`
	Description string         `json:"description"`
	Content     datatypes.JSON `gorm:"type:jsonb" json:"content"`
	OwnerID     uuid.UUID      `json:"owner_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Owner User `gorm:"foreignKey:OwnerID" json:"owner"`
}
