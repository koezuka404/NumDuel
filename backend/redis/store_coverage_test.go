package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func closedStore(t *testing.T) *Store {
	t.Helper()
	store, mr := newTestStore(t)
	mr.Close()
	return store
}

func TestSetTurnRedisError(t *testing.T) {
	store := closedStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	err := store.SetTurn(ctx, uuid.New(), 1, uuid.New(), now, now.Add(time.Minute))
	if err == nil {
		t.Fatal("expected SetTurn redis error")
	}
}

func TestRemainingSecondsRedisError(t *testing.T) {
	store := closedStore(t)
	_, err := store.RemainingSeconds(context.Background(), uuid.New(), time.Now())
	if err == nil {
		t.Fatal("expected Get redis error")
	}
}

func TestGetTurnRedisError(t *testing.T) {
	store := closedStore(t)
	_, err := store.GetTurn(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected GetTurn redis error")
	}
}

func TestGetForceLogoutBeforeRedisError(t *testing.T) {
	store := closedStore(t)
	_, err := store.GetForceLogoutBefore(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected Get redis error")
	}
}

func TestGetBackupStatusRedisError(t *testing.T) {
	store := closedStore(t)
	_, err := store.GetBackupStatus(context.Background())
	if err == nil {
		t.Fatal("expected GetBackupStatus redis error")
	}
}

func TestSetBackupStatusRedisError(t *testing.T) {
	store := closedStore(t)
	err := store.SetBackupStatus(context.Background(), "failed", time.Now().UTC())
	if err == nil {
		t.Fatal("expected SetBackupStatus redis error")
	}
}

func TestRevokeRedisError(t *testing.T) {
	store := closedStore(t)
	if err := store.Revoke(context.Background(), "jti", time.Minute); err == nil {
		t.Fatal("expected Revoke redis error")
	}
}

func TestIsRevokedRedisError(t *testing.T) {
	store := closedStore(t)
	_, err := store.IsRevoked(context.Background(), "jti")
	if err == nil {
		t.Fatal("expected IsRevoked redis error")
	}
}

func TestSetUserRedisError(t *testing.T) {
	store := closedStore(t)
	if err := store.SetUser(context.Background(), uuid.New(), "c1", time.Minute); err == nil {
		t.Fatal("expected SetUser redis error")
	}
}

func TestDeleteUserRedisError(t *testing.T) {
	store := closedStore(t)
	if err := store.DeleteUser(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected DeleteUser redis error")
	}
}

func TestDeleteTurnRedisError(t *testing.T) {
	store := closedStore(t)
	if err := store.DeleteTurn(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected DeleteTurn redis error")
	}
}

func TestSetForceLogoutBeforeRedisError(t *testing.T) {
	store := closedStore(t)
	if err := store.SetForceLogoutBefore(context.Background(), uuid.New(), time.Now().UTC()); err == nil {
		t.Fatal("expected SetForceLogoutBefore redis error")
	}
}

func TestListExpiredTurnsSkipsBrokenTurnKeys(t *testing.T) {
	store, mr := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	brokenID := uuid.New()
	mr.Set(turnKey(brokenID), `{`)

	expiredGame := uuid.New()
	expiredPlayer := uuid.New()
	if err := store.SetTurn(ctx, expiredGame, 1, expiredPlayer, now.Add(-2*time.Minute), now.Add(-time.Second)); err != nil {
		t.Fatalf("set turn: %v", err)
	}

	entries, err := store.ListExpiredTurns(ctx, now)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 1 || entries[0].GameID != expiredGame {
		t.Fatalf("entries: %+v", entries)
	}
}

func TestListExpiredTurnsPaginatedScan(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	expiredPlayer := uuid.New()

	for i := 0; i < 105; i++ {
		gameID := uuid.New()
		if err := store.SetTurn(ctx, gameID, 1, expiredPlayer, now.Add(-2*time.Minute), now.Add(-time.Second)); err != nil {
			t.Fatalf("set turn %d: %v", i, err)
		}
	}

	entries, err := store.ListExpiredTurns(ctx, now)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 105 {
		t.Fatalf("expected 105 expired turns, got %d", len(entries))
	}
}

func TestParseTurnKeyGameIDInvalidPrefixSuffix(t *testing.T) {
	if _, err := parseTurnKeyGameID("game:only"); err == nil {
		t.Fatal("expected invalid key")
	}
	if _, err := parseTurnKeyGameID("wrong:" + uuid.New().String() + ":turn"); err == nil {
		t.Fatal("expected prefix mismatch")
	}
}

func TestSetTurnMarshalError(t *testing.T) {
	orig := marshalJSON
	t.Cleanup(func() { marshalJSON = orig })
	marshalJSON = func(any) ([]byte, error) { return nil, errors.New("marshal failed") }

	store, _ := newTestStore(t)
	now := time.Now().UTC()
	err := store.SetTurn(context.Background(), uuid.New(), 1, uuid.New(), now, now.Add(time.Minute))
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestSetBackupStatusMarshalError(t *testing.T) {
	orig := marshalJSON
	t.Cleanup(func() { marshalJSON = orig })
	marshalJSON = func(any) ([]byte, error) { return nil, errors.New("marshal failed") }

	store, _ := newTestStore(t)
	err := store.SetBackupStatus(context.Background(), "ok", time.Now().UTC())
	if err == nil {
		t.Fatal("expected marshal error")
	}
}
