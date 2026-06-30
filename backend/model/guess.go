package model

import (
	"time"

	"github.com/google/uuid"
)

type Guess struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey"`
	GameID       uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:uq_guesses_game_turn_player;index:idx_guesses_game_turn,priority:1"`
	PlayerID     uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:uq_guesses_game_turn_player"`
	Turn         int           `gorm:"not null;uniqueIndex:uq_guesses_game_turn_player;index:idx_guesses_game_turn,priority:2"`
	GuessNumber  string        `gorm:"size:4;not null"`
	DigitResults []DigitResult `gorm:"type:jsonb;serializer:json"`
	HitCount     int           `gorm:"not null"`
	IsAuto       bool          `gorm:"not null;default:false"`
	CreatedAt    time.Time     `gorm:"not null"`
	UpdatedAt    time.Time     `gorm:"not null"`
}

func (Guess) TableName() string { return "guesses" }
