package model

import (
	"time"

	"github.com/google/uuid"
)

type MatchHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	GameID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	WinnerID       uuid.UUID `gorm:"type:uuid;not null;index"`
	LoserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	WinnerUsername string    `gorm:"size:50;not null"`
	LoserUsername  string    `gorm:"size:50;not null"`
	FinishedAt     time.Time `gorm:"not null"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (MatchHistory) TableName() string { return "match_histories" }
