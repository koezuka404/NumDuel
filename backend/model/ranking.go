package model

import (
	"time"

	"github.com/google/uuid"
)

type Ranking struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Rank      int       `gorm:"not null;index"`
	Username  string    `gorm:"size:50;not null"`
	WinCount  int       `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Ranking) TableName() string { return "rankings" }
