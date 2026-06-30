package model

import (
	"time"

	"github.com/google/uuid"
)

type Game struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Status              GameStatus `gorm:"size:20;not null;index"`
	Player1ID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	Player2ID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	Player1Secret       string     `gorm:"size:255"`
	Player2Secret       string     `gorm:"size:255"`
	CurrentTurnPlayerID *uuid.UUID `gorm:"type:uuid"`
	CurrentTurn         int        `gorm:"not null;default:1"`
	WinnerID            *uuid.UUID `gorm:"type:uuid"`
	StartedAt           *time.Time
	FinishedAt          *time.Time
	CreatedAt           time.Time `gorm:"not null"`
	UpdatedAt           time.Time `gorm:"not null"`
}

func (Game) TableName() string { return "games" }

func (g Game) IsParticipant(userID uuid.UUID) bool {
	return g.Player1ID == userID || g.Player2ID == userID
}

func (g Game) IsCurrentTurn(userID uuid.UUID) bool {
	if g.CurrentTurnPlayerID == nil {
		return false
	}
	return *g.CurrentTurnPlayerID == userID
}

func (g Game) CanGuess(userID uuid.UUID) bool {
	return g.Status == GameStatusInProgress && g.IsCurrentTurn(userID)
}

func (g Game) BothSecretsSet() bool {
	return g.Player1Secret != "" && g.Player2Secret != ""
}

func (g Game) IsFinished() bool {
	return g.Status == GameStatusFinished
}
