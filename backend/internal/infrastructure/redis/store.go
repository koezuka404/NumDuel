// Redis: JWT 失効・WS セッション・ゲームロック・ターン期限。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/numduel/numduel/internal/domain"
)

type Store struct {
	client *goredis.Client
}

var (
	_ domain.JWTRevoker       = (*Store)(nil)
	_ domain.WSSessionStore   = (*Store)(nil)
	_ domain.GameLockStore    = (*Store)(nil)
	_ domain.TurnStore        = (*Store)(nil)
	_ domain.ForceLogoutStore = (*Store)(nil)
)

type turnValue struct {
	Turn      int       `json:"turn"`
	PlayerID  string    `json:"playerId"`
	StartedAt time.Time `json:"startedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func Open(url string) (*Store, error) {
	opt, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := goredis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Store{client: client}, nil
}

func (s *Store) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}

func (s *Store) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return s.client.Set(ctx, "jwt:revoked:"+jti, "1", ttl).Err()
}

func (s *Store) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := s.client.Exists(ctx, "jwt:revoked:"+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Store) SetUser(ctx context.Context, userID uuid.UUID, connectionID string, ttl time.Duration) error {
	return s.client.Set(ctx, "ws:user:"+userID.String(), connectionID, ttl).Err()
}

func (s *Store) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.client.Del(ctx, "ws:user:"+userID.String()).Err()
}

func (s *Store) GetForceLogoutBefore(ctx context.Context, userID uuid.UUID) (time.Time, error) {
	raw, err := s.client.Get(ctx, "user:"+userID.String()+":force_logout_before").Result()
	if err == goredis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	sec, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0).UTC(), nil
}

func (s *Store) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := s.client.SetNX(ctx, key, "1", ttl).Result()
	return ok, err
}

func SecretLockKey(gameID, playerID uuid.UUID) string {
	return fmt.Sprintf("game:%s:player:%s:secret_lock", gameID, playerID)
}

func GuessLockKey(gameID, playerID uuid.UUID) string {
	return fmt.Sprintf("game:%s:player:%s:guess_lock", gameID, playerID)
}

func turnKey(gameID uuid.UUID) string {
	return "game:" + gameID.String() + ":turn"
}

func (s *Store) SetTurn(ctx context.Context, gameID uuid.UUID, turn int, playerID uuid.UUID, startedAt, expiresAt time.Time) error {
	v := turnValue{
		Turn: turn, PlayerID: playerID.String(),
		StartedAt: startedAt.UTC(), ExpiresAt: expiresAt.UTC(),
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ttl := time.Until(expiresAt)
	if ttl < time.Second {
		ttl = time.Second
	}
	return s.client.Set(ctx, turnKey(gameID), b, ttl).Err()
}

func (s *Store) RemainingSeconds(ctx context.Context, gameID uuid.UUID, now time.Time) (int, error) {
	raw, err := s.client.Get(ctx, turnKey(gameID)).Result()
	if err == goredis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var v turnValue
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return 0, err
	}
	sec := int(v.ExpiresAt.Sub(now).Seconds())
	if sec < 0 {
		return 0, nil
	}
	return sec, nil
}

func (s *Store) DeleteTurn(ctx context.Context, gameID uuid.UUID) error {
	return s.client.Del(ctx, turnKey(gameID)).Err()
}
