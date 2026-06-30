// ドメイン層の外部サービスインターフェース。Infrastructure が実装する。
package model

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

type AccessTokenIssuer interface {
	Issue(userID uuid.UUID, role Role, now time.Time) (string, error)
}

type RefreshTokenPair struct {
	Plaintext string // Cookie / レスポンス用（DB には保存しない）
	Hash      string // refresh_tokens.token_hash に保存
}

type RefreshTokenGenerator interface {
	Generate() (RefreshTokenPair, error)
	Hash(plaintext string) string
}

type JWTRevoker interface {
	Revoke(ctx context.Context, jti string, ttl time.Duration) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

type WSSessionStore interface {
	SetUser(ctx context.Context, userID uuid.UUID, connectionID string, ttl time.Duration) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

// SecretHasher は秘密数字の位置別 HMAC 生成・照合。
type SecretHasher interface {
	Hash(digits [4]int, gameID uuid.UUID, playerSlot int) (string, error)
	Verify(storedHash string, guess GuessNumber, gameID uuid.UUID, opponentSlot int) ([4]DigitResult, error)
}

// GameLockStore は Redis 連打防止ロック。
type GameLockStore interface {
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// TurnInfo は Redis game:{gameId}:turn の内容。
type TurnInfo struct {
	Turn      int
	PlayerID  uuid.UUID
	StartedAt time.Time
	ExpiresAt time.Time
}

// GuessNumberGenerator はタイムアウト自動予想の 4 桁生成。
type GuessNumberGenerator interface {
	GenerateGuessNumber() (GuessNumber, error)
}

// TurnStore はターン期限（game:{gameId}:turn）の管理。
type TurnStore interface {
	SetTurn(ctx context.Context, gameID uuid.UUID, turn int, playerID uuid.UUID, startedAt, expiresAt time.Time) error
	GetTurn(ctx context.Context, gameID uuid.UUID) (*TurnInfo, error)
	RemainingSeconds(ctx context.Context, gameID uuid.UUID, now time.Time) (int, error)
	DeleteTurn(ctx context.Context, gameID uuid.UUID) error
}

// ForceLogoutStore は user:{userId}:force_logout_before の管理。
type ForceLogoutStore interface {
	GetForceLogoutBefore(ctx context.Context, userID uuid.UUID) (time.Time, error)
}

// BackupStatus は backup:status Redis キーの値。
type BackupStatus struct {
	LastSyncedAt *time.Time
	Status       string // ok / error
}

// BackupStatusStore はバックアップ同期状態の参照。
type BackupStatusStore interface {
	GetBackupStatus(ctx context.Context) (*BackupStatus, error)
}

// EventNotifier は DB コミット後の WebSocket 通知用（現状 no-op 実装）。
type EventNotifier interface {
	SendToUser(ctx context.Context, userID uuid.UUID, eventType string, payload map[string]any) error
}
