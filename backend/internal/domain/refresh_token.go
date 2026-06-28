// RefreshToken エンティティ。平文は Cookie のみ、DB にはハッシュを保存。
package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string // SHA-256 hex
	FamilyID  uuid.UUID // 同一ログインセッションのローテーション単位
	Status    RefreshTokenStatus
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

// IsActive は status=active かつ期限内。
func (t *RefreshToken) IsActive(now time.Time) bool {
	return t != nil && t.Status == RefreshTokenActive && now.Before(t.ExpiresAt)
}

// Revoke は status を revoked に更新する。
func (t *RefreshToken) Revoke(now time.Time) {
	t.Status = RefreshTokenRevoked
	t.RevokedAt = &now
	t.UpdatedAt = now
}
