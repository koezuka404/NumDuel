package model

import (
	"time"

	"github.com/google/uuid"
)

type WSConnectionLog struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null"`
	ConnectionID   string     `gorm:"size:64;not null"`
	ConnectedAt    time.Time  `gorm:"not null"`
	DisconnectedAt *time.Time
}

func (WSConnectionLog) TableName() string { return "ws_connection_logs" }
