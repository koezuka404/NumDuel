package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID                uuid.UUID          `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID          `gorm:"type:uuid;not null"`
	TokenHash         string             `gorm:"size:255;not null;uniqueIndex"`
	FamilyID          uuid.UUID          `gorm:"type:uuid;not null"`
	Status            RefreshTokenStatus `gorm:"size:20;not null"`
	ExpiresAt         time.Time          `gorm:"not null"`
	RevokedAt         *time.Time
	ReplacedByTokenID *uuid.UUID         `gorm:"type:uuid"`
	CreatedAt         time.Time          `gorm:"not null"`
	UpdatedAt         time.Time          `gorm:"not null"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

func (t RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil || t.Status == RefreshTokenRevoked
}

func (t RefreshToken) IsExpired(now time.Time) bool {
	return !t.ExpiresAt.After(now)
}

func (t RefreshToken) IsActive(now time.Time) bool {
	return !t.IsRevoked() && !t.IsExpired(now)
}

func NewRefreshToken(userID uuid.UUID, hash string, familyID uuid.UUID, expiresAt, now time.Time) RefreshToken {
	return RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hash,
		FamilyID:  familyID,
		Status:    RefreshTokenActive,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
