package model

import (
	"time"

	"github.com/google/uuid"
)

type LoginLog struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID   `gorm:"type:uuid;not null"`
	Action    LoginAction `gorm:"size:20;not null"`
	CreatedAt time.Time   `gorm:"not null"`
	UpdatedAt time.Time   `gorm:"not null"`
}

func (LoginLog) TableName() string { return "login_logs" }
