package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

func newTestStore(t *testing.T) (*Store, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewStore(rdb), mr
}

func TestNewStoreNilClient(t *testing.T) {
	if NewStore(nil) != nil {
		t.Fatal("expected nil store")
	}
}

func TestJWTRevokeAndCheck(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()

	if err := store.Revoke(ctx, "", time.Minute); err != nil {
		t.Fatalf("empty jti: %v", err)
	}
	if err := store.Revoke(ctx, "jti-1", 0); err != nil {
		t.Fatalf("zero ttl: %v", err)
	}
	if err := store.Revoke(ctx, "jti-1", time.Minute); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	ok, err := store.IsRevoked(ctx, "jti-1")
	if err != nil || !ok {
		t.Fatalf("revoked: ok=%v err=%v", ok, err)
	}
	ok, err = store.IsRevoked(ctx, "other")
	if err != nil || ok {
		t.Fatalf("not revoked: ok=%v err=%v", ok, err)
	}
}

func TestWSSessionUser(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	userID := uuid.New()

	if err := store.SetUser(ctx, userID, "conn-1", 0); err != nil {
		t.Fatalf("set user default ttl: %v", err)
	}
	if err := store.DeleteUser(ctx, userID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
}

func TestAcquireLock(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()

	ok, err := store.AcquireLock(ctx, "lock:test", time.Second)
	if err != nil || !ok {
		t.Fatalf("first acquire: ok=%v err=%v", ok, err)
	}
	ok, err = store.AcquireLock(ctx, "lock:test", time.Second)
	if err != nil || ok {
		t.Fatalf("second acquire: ok=%v err=%v", ok, err)
	}
	ok, err = store.AcquireLock(ctx, "lock:zero", 0)
	if err != nil || !ok {
		t.Fatalf("zero ttl acquire: ok=%v err=%v", ok, err)
	}
}

func TestTurnLifecycle(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	gameID := uuid.New()
	playerID := uuid.New()
	started := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	expires := started.Add(30 * time.Second)

	if err := store.SetTurn(ctx, gameID, 1, playerID, started, expires); err != nil {
		t.Fatalf("set turn: %v", err)
	}
	info, err := store.GetTurn(ctx, gameID)
	if err != nil || info == nil || info.PlayerID != playerID || info.Turn != 1 {
		t.Fatalf("get turn: %+v err=%v", info, err)
	}
	remaining, err := store.RemainingSeconds(ctx, gameID, started.Add(10*time.Second))
	if err != nil || remaining != 20 {
		t.Fatalf("remaining: %d err=%v", remaining, err)
	}
	remaining, err = store.RemainingSeconds(ctx, gameID, expires.Add(time.Minute))
	if err != nil || remaining != 0 {
		t.Fatalf("expired remaining: %d err=%v", remaining, err)
	}
	if _, err := store.RemainingSeconds(ctx, uuid.New(), time.Now()); err != nil {
		t.Fatalf("missing turn: %v", err)
	}
	if err := store.DeleteTurn(ctx, gameID); err != nil {
		t.Fatalf("delete turn: %v", err)
	}
	if info, err := store.GetTurn(ctx, gameID); err != nil || info != nil {
		t.Fatalf("after delete: %+v err=%v", info, err)
	}
}

func TestSetTurnMinimumTTL(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	if err := store.SetTurn(ctx, uuid.New(), 1, uuid.New(), now, now.Add(100*time.Millisecond)); err != nil {
		t.Fatalf("set short turn: %v", err)
	}
}

func TestGetTurnInvalidPayload(t *testing.T) {
	store, mr := newTestStore(t)
	ctx := context.Background()
	gameID := uuid.New()
	mr.Set(turnKey(gameID), `{`)

	if _, err := store.GetTurn(ctx, gameID); err == nil {
		t.Fatal("expected json error")
	}

	mr.Set(turnKey(gameID), `{"turn":1,"playerId":"not-uuid","startedAt":"2026-07-01T00:00:00Z","expiresAt":"2026-07-01T00:00:30Z"}`)
	if _, err := store.GetTurn(ctx, gameID); err == nil {
		t.Fatal("expected uuid parse error")
	}
}

func TestListExpiredTurns(t *testing.T) {
	store, mr := newTestStore(t)
	ctx := context.Background()
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	activeGame := uuid.New()
	activePlayer := uuid.New()
	if err := store.SetTurn(ctx, activeGame, 1, activePlayer, now.Add(-time.Minute), now.Add(time.Minute)); err != nil {
		t.Fatalf("active turn: %v", err)
	}

	expiredGame := uuid.New()
	expiredPlayer := uuid.New()
	if err := store.SetTurn(ctx, expiredGame, 2, expiredPlayer, now.Add(-2*time.Minute), now.Add(-time.Second)); err != nil {
		t.Fatalf("expired turn: %v", err)
	}

	mr.Set("game:not-a-uuid:turn", "x")
	mr.Set("bad-key", "x")

	entries, err := store.ListExpiredTurns(ctx, now)
	if err != nil {
		t.Fatalf("list expired: %v", err)
	}
	if len(entries) != 1 || entries[0].GameID != expiredGame || entries[0].PlayerID != expiredPlayer {
		t.Fatalf("entries: %+v", entries)
	}
}

func TestForceLogoutBefore(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	userID := uuid.New()

	before, err := store.GetForceLogoutBefore(ctx, userID)
	if err != nil || !before.IsZero() {
		t.Fatalf("empty: %v before=%v", err, before)
	}

	ts := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	if err := store.SetForceLogoutBefore(ctx, userID, ts); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := store.GetForceLogoutBefore(ctx, userID)
	if err != nil || !got.Equal(ts) {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestGetForceLogoutBeforeInvalidValue(t *testing.T) {
	store, mr := newTestStore(t)
	ctx := context.Background()
	userID := uuid.New()
	mr.Set(forceLogoutKey(userID), "not-a-number")

	if _, err := store.GetForceLogoutBefore(ctx, userID); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParseTurnKeyGameID(t *testing.T) {
	id, err := parseTurnKeyGameID("game:" + uuid.New().String() + ":turn")
	if err != nil || id == uuid.Nil {
		t.Fatalf("valid key: %v id=%v", err, id)
	}
	for _, key := range []string{"", "game:x", "game:bad", "other:uuid:turn"} {
		if _, err := parseTurnKeyGameID(key); err == nil {
			t.Fatalf("expected error for %q", key)
		}
	}
}

func TestRemainingSecondsInvalidJSON(t *testing.T) {
	store, mr := newTestStore(t)
	gameID := uuid.New()
	mr.Set(turnKey(gameID), `{`)
	if _, err := store.RemainingSeconds(context.Background(), gameID, time.Now()); err == nil {
		t.Fatal("expected json error")
	}
}

func TestAcquireLockRedisError(t *testing.T) {
	store, mr := newTestStore(t)
	mr.Close()
	if _, err := store.AcquireLock(context.Background(), "lock:dead", time.Second); err == nil {
		t.Fatal("expected redis error")
	}
}

func TestListExpiredTurnsScanError(t *testing.T) {
	store, mr := newTestStore(t)
	mr.Close()
	if _, err := store.ListExpiredTurns(context.Background(), time.Now()); err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSetBackupStatusWithoutTimestamp(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	if err := store.SetBackupStatus(ctx, "failed", time.Time{}); err != nil {
		t.Fatalf("SetBackupStatus: %v", err)
	}
	st, err := store.GetBackupStatus(ctx)
	if err != nil || st.Status != "failed" {
		t.Fatalf("status: %+v err=%v", st, err)
	}
}

func TestBackupStatus(t *testing.T) {
	store, mr := newTestStore(t)
	ctx := context.Background()

	st, err := store.GetBackupStatus(ctx)
	if err != nil || st.Status != "ok" {
		t.Fatalf("default status: %+v err=%v", st, err)
	}

	synced := time.Date(2026, 7, 4, 9, 0, 0, 0, time.UTC)
	if err := store.SetBackupStatus(ctx, "synced", synced); err != nil {
		t.Fatalf("set status: %v", err)
	}
	st, err = store.GetBackupStatus(ctx)
	if err != nil || st.Status != "synced" || st.LastSyncedAt == nil || !st.LastSyncedAt.Equal(synced) {
		t.Fatalf("status: %+v err=%v", st, err)
	}

	mr.Set(backupStatusKey(), `{`)
	if _, err := store.GetBackupStatus(ctx); err == nil {
		t.Fatal("expected json error")
	}

	raw, _ := json.Marshal(map[string]string{
		"status": "failed", "lastSyncedAt": "not-rfc3339",
	})
	mr.Set(backupStatusKey(), string(raw))
	if _, err := store.GetBackupStatus(ctx); err == nil {
		t.Fatal("expected date parse error")
	}

	mr.Set(backupStatusKey(), `{"status":""}`)
	st, err = store.GetBackupStatus(ctx)
	if err != nil || st.Status != "ok" {
		t.Fatalf("empty status defaults ok: %+v err=%v", st, err)
	}
}
