// ドメイン層の外部サービスインターフェース。Infrastructure が実装する。
package domain

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
}

type JWTRevoker interface {
	Revoke(ctx context.Context, jti string, ttl time.Duration) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

type WSSessionStore interface {
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}
