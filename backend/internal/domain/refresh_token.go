package domain

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken はリフレッシュトークンのメタデータ Entity（仕様書 Login/Refresh UseCase）。
//
// 保管方針:
//   - 平文トークンは HttpOnly Cookie のみで運搬（Request Body 不使用）
//   - DB には SHA-256 ハッシュ（token_hash）のみ保存
//   - family_id … 同一ログインセッション内のローテーション単位。盗用検出時に一括失効
//
// テーブル: refresh_tokens（仕様書 認証章, 13.4）
type RefreshToken struct {
	ID        uuid.UUID          // PK
	UserID    uuid.UUID          // FK → users
	TokenHash string             // SHA-256(平文) hex。UNIQUE
	FamilyID  uuid.UUID          // ログインセッション単位の ID
	Status    RefreshTokenStatus // active / revoked
	ExpiresAt time.Time          // REFRESH_TOKEN_EXPIRY_DAYS 後
	RevokedAt *time.Time         // 失効日時（revoked 時）
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewRefreshToken は LoginUseCase / RefreshTokenUseCase が新規発行時に使用。
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

// IsActive は現在有効なトークンか（status=active かつ期限内）。
func (t *RefreshToken) IsActive(now time.Time) bool {
	return t != nil && t.Status == RefreshTokenActive && now.Before(t.ExpiresAt)
}

// Revoke はトークンを失効させる。Logout / ローテーション / family 一括失効で使用。
func (t *RefreshToken) Revoke(now time.Time) {
	t.Status = RefreshTokenRevoked
	t.RevokedAt = &now
	t.UpdatedAt = now
}
