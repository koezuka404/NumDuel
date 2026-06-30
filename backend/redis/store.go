// Redis 補助ストア
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/numduel/numduel/model"
)

const (
	lockAcquireScript = `
if redis.call("SET", KEYS[1], "1", "NX", "EX", ARGV[1]) then
  return 1
end
return 0
`
	forceLogoutTTL = 30 * 24 * time.Hour
)

type Store struct {
	rdb *goredis.Client
}

var (
	_ model.JWTRevoker        = (*Store)(nil)
	_ model.WSSessionStore    = (*Store)(nil)
	_ model.GameLockStore     = (*Store)(nil)
	_ model.TurnStore         = (*Store)(nil)
	_ model.ForceLogoutStore  = (*Store)(nil)
	_ model.BackupStatusStore = (*Store)(nil)
)

func NewStore(rdb *goredis.Client) *Store {
	if rdb == nil {
		return nil
	}
	return &Store{rdb: rdb}
}

func (s *Store) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" || ttl <= 0 {
		return nil
	}
	return s.rdb.Set(ctx, jwtRevokedKey(jti), "1", ttl).Err()
}

func (s *Store) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := s.rdb.Exists(ctx, jwtRevokedKey(jti)).Result()
	return n > 0, err
}

func (s *Store) SetUser(ctx context.Context, userID uuid.UUID, connectionID string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return s.rdb.Set(ctx, wsUserKey(userID), connectionID, ttl).Err()
}

func (s *Store) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.rdb.Del(ctx, wsUserKey(userID)).Err()
}

func (s *Store) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	sec := int(ttl.Seconds())
	if sec <= 0 {
		sec = 1
	}
	res, err := s.rdb.Eval(ctx, lockAcquireScript, []string{key}, sec).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

type turnPayload struct {
	Turn      int       `json:"turn"`
	PlayerID  string    `json:"playerId"`
	StartedAt time.Time `json:"startedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (s *Store) SetTurn(ctx context.Context, gameID uuid.UUID, turn int, playerID uuid.UUID, startedAt, expiresAt time.Time) error {
	payload := turnPayload{
		Turn:      turn,
		PlayerID:  playerID.String(),
		StartedAt: startedAt.UTC(),
		ExpiresAt: expiresAt.UTC(),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ttl := expiresAt.Sub(startedAt)
	if ttl < time.Second {
		ttl = time.Second
	}
	return s.rdb.Set(ctx, turnKey(gameID), b, ttl).Err()
}

func (s *Store) RemainingSeconds(ctx context.Context, gameID uuid.UUID, now time.Time) (int, error) {
	val, err := s.rdb.Get(ctx, turnKey(gameID)).Bytes()
	if err == goredis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var p turnPayload
	if err := json.Unmarshal(val, &p); err != nil {
		return 0, err
	}
	remaining := int(p.ExpiresAt.Sub(now.UTC()).Seconds())
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

func (s *Store) DeleteTurn(ctx context.Context, gameID uuid.UUID) error {
	return s.rdb.Del(ctx, turnKey(gameID)).Err()
}

func (s *Store) GetTurn(ctx context.Context, gameID uuid.UUID) (*model.TurnInfo, error) {
	val, err := s.rdb.Get(ctx, turnKey(gameID)).Bytes()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var p turnPayload
	if err := json.Unmarshal(val, &p); err != nil {
		return nil, err
	}
	playerID, err := uuid.Parse(p.PlayerID)
	if err != nil {
		return nil, err
	}
	return &model.TurnInfo{
		Turn:      p.Turn,
		PlayerID:  playerID,
		StartedAt: p.StartedAt.UTC(),
		ExpiresAt: p.ExpiresAt.UTC(),
	}, nil
}

// ExpiredTurnEntry は期限切れターン 1 件
type ExpiredTurnEntry struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID
}

// ListExpiredTurns は game:*:turn のうち expiresAt <= now のものを返す
func (s *Store) ListExpiredTurns(ctx context.Context, now time.Time) ([]ExpiredTurnEntry, error) {
	now = now.UTC()
	var (
		out    []ExpiredTurnEntry
		cursor uint64
	)
	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, "game:*:turn", 100).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			gameID, err := parseTurnKeyGameID(key)
			if err != nil {
				continue
			}
			info, err := s.GetTurn(ctx, gameID)
			if err != nil || info == nil {
				continue
			}
			if !info.ExpiresAt.After(now) {
				out = append(out, ExpiredTurnEntry{GameID: gameID, PlayerID: info.PlayerID})
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return out, nil
}

func parseTurnKeyGameID(key string) (uuid.UUID, error) {
	const prefix = "game:"
	const suffix = ":turn"
	if len(key) <= len(prefix)+len(suffix) {
		return uuid.Nil, fmt.Errorf("invalid turn key")
	}
	if key[:len(prefix)] != prefix || key[len(key)-len(suffix):] != suffix {
		return uuid.Nil, fmt.Errorf("invalid turn key")
	}
	return uuid.Parse(key[len(prefix) : len(key)-len(suffix)])
}

func (s *Store) GetForceLogoutBefore(ctx context.Context, userID uuid.UUID) (time.Time, error) {
	val, err := s.rdb.Get(ctx, forceLogoutKey(userID)).Result()
	if err == goredis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	sec, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0).UTC(), nil
}

// SetForceLogoutBefore は AutoLogoutWorker / 管理操作向けTTL 30 日固定
func (s *Store) SetForceLogoutBefore(ctx context.Context, userID uuid.UUID, t time.Time) error {
	return s.rdb.Set(ctx, forceLogoutKey(userID), strconv.FormatInt(t.UTC().Unix(), 10), forceLogoutTTL).Err()
}

func (s *Store) GetBackupStatus(ctx context.Context) (*model.BackupStatus, error) {
	val, err := s.rdb.Get(ctx, backupStatusKey()).Bytes()
	if err == goredis.Nil {
		return &model.BackupStatus{Status: "ok"}, nil
	}
	if err != nil {
		return nil, err
	}
	var raw struct {
		Status       string `json:"status"`
		LastSyncedAt string `json:"lastSyncedAt"`
	}
	if err := json.Unmarshal(val, &raw); err != nil {
		return nil, err
	}
	out := &model.BackupStatus{Status: raw.Status}
	if raw.LastSyncedAt != "" {
		t, err := time.Parse(time.RFC3339, raw.LastSyncedAt)
		if err != nil {
			return nil, err
		}
		utc := t.UTC()
		out.LastSyncedAt = &utc
	}
	if out.Status == "" {
		out.Status = "ok"
	}
	return out, nil
}

// SetBackupStatus は BackupWorker 成功/失敗時に更新する
func (s *Store) SetBackupStatus(ctx context.Context, status string, lastSyncedAt time.Time) error {
	payload, err := json.Marshal(map[string]string{
		"status":       status,
		"lastSyncedAt": lastSyncedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, backupStatusKey(), payload, 0).Err()
}

func jwtRevokedKey(jti string) string {
	return "jwt:revoked:" + jti
}

func wsUserKey(userID uuid.UUID) string {
	return fmt.Sprintf("ws:user:%s", userID)
}

func forceLogoutKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:force_logout_before", userID)
}

func turnKey(gameID uuid.UUID) string {
	return fmt.Sprintf("game:%s:turn", gameID)
}

func backupStatusKey() string {
	return "backup:status"
}
