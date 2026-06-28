// Redis 操作のプレースホルダ。JWT 失効・WS セッション管理は Redis 実装後に差し替える。
package redis

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/internal/domain"
)

type Store struct{}

var (
	_ domain.JWTRevoker     = (*Store)(nil)
	_ domain.WSSessionStore = (*Store)(nil)
)

func NewStore() *Store { return &Store{} }

// Revoke は jwt:revoked:{jti} を Redis に SET する（現状 no-op）。
func (s *Store) Revoke(ctx context.Context, jti string, ttl time.Duration) error { return nil }

func (s *Store) IsRevoked(ctx context.Context, jti string) (bool, error) { return false, nil }

// DeleteUser は ws:user:{userId} を削除する（現状 no-op）。
func (s *Store) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}
