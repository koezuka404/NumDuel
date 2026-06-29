package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ActivityLog は activity_logs テーブルのレコード。
type ActivityLog struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID      `gorm:"type:uuid"`
	LogType   string          `gorm:"size:50;not null;index:idx_activity_logs_type_created,priority:1"`
	Detail    json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt time.Time       `gorm:"not null;index:idx_activity_logs_type_created,priority:2"`
	UpdatedAt time.Time       `gorm:"not null"`
}

func (ActivityLog) TableName() string { return "activity_logs" }

// LoginLog は login_logs テーブルのレコード。
type LoginLog struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID   `gorm:"type:uuid;not null"`
	Action    LoginAction `gorm:"size:20;not null"`
	CreatedAt time.Time   `gorm:"not null"`
	UpdatedAt time.Time   `gorm:"not null"`
}

func (LoginLog) TableName() string { return "login_logs" }

// WSConnectionLog は ws_connection_logs テーブルのレコード。
type WSConnectionLog struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null"`
	ConnectionID   string     `gorm:"size:64;not null"`
	ConnectedAt    time.Time  `gorm:"not null"`
	DisconnectedAt *time.Time
}

func (WSConnectionLog) TableName() string { return "ws_connection_logs" }
