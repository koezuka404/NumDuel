package model

import (
	"time"

	"github.com/google/uuid"
)

// MatchingQueueEntry は matching_queue テーブルのレコード
type MatchingQueueEntry struct {
	ID        uuid.UUID           `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID           `gorm:"type:uuid;not null"`
	Status    MatchingQueueStatus `gorm:"size:20;not null;index:idx_matching_queue_status_created,priority:1"`
	CreatedAt time.Time           `gorm:"not null;index:idx_matching_queue_status_created,priority:2"`
}

func (MatchingQueueEntry) TableName() string { return "matching_queue" }
