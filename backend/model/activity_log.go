package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID      `gorm:"type:uuid"`
	LogType   string          `gorm:"size:50;not null;index:idx_activity_logs_type_created,priority:1"`
	Detail    json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt time.Time       `gorm:"not null;index:idx_activity_logs_type_created,priority:2"`
	UpdatedAt time.Time       `gorm:"not null"`
}

func (ActivityLog) TableName() string { return "activity_logs" }
