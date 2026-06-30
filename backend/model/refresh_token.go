// RefreshToken エンティティ平文は Cookie のみ、DB にはハッシュを保存
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
	User              User               `gorm:"foreignKey:UserID;references:ID"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

func NewRefreshToken(
	userID uuid.UUID,
	tokenHash string,
	familyID uuid.UUID,
	expiresAt time.Time,
	now time.Time,
) RefreshToken {
	return RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		FamilyID:  familyID,
		Status:    RefreshTokenActive,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsActive は status=active かつ期限内
func (t *RefreshToken) IsActive(now time.Time) bool {
	return t != nil && t.Status == RefreshTokenActive && now.Before(t.ExpiresAt)
}

// Revoke は status を revoked に更新する
func (t *RefreshToken) Revoke(now time.Time) {
	t.Status = RefreshTokenRevoked
	t.RevokedAt = &now
	t.UpdatedAt = now
}
