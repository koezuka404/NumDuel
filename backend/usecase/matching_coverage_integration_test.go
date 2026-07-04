package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestMatchingCancel(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	if _, err := match.Start(context.Background(), user.ID); err != nil {
		t.Fatalf("start: %v", err)
	}
	out, err := match.Cancel(context.Background(), user.ID)
	if err != nil || out.Status != "cancelled" {
		t.Fatalf("cancel: %+v err=%v", out, err)
	}
	status, err := match.Status(context.Background(), user.ID)
	if err != nil || status.Status != "idle" {
		t.Fatalf("status after cancel: %+v err=%v", status, err)
	}
}

func TestMatchingStartAlreadyWaiting(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	if _, err := match.Start(context.Background(), user.ID); err != nil {
		t.Fatalf("start: %v", err)
	}
	_, err := match.Start(context.Background(), user.ID)
	if !errors.Is(err, usecase.ErrAlreadyInMatching) {
		t.Fatalf("already waiting: %v", err)
	}
}

func TestMatchingMasterForbidden(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	_, err := match.Start(context.Background(), master.ID)
	if !errors.Is(err, usecase.ErrForbidden) {
		t.Fatalf("master forbidden: %v", err)
	}
}

func TestMatchingStatusMatched(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	status, err := match.Status(context.Background(), a.ID)
	if err != nil || status.Status != "matched" || status.GameID == nil || *status.GameID != gameID {
		t.Fatalf("matched status: %+v err=%v", status, err)
	}
}

func TestMatchingStartWithoutNotifier(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := usecase.NewMatchingUseCase(repos, nil)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")

	if _, err := match.Start(context.Background(), a.ID); err != nil {
		t.Fatalf("alice start: %v", err)
	}
	if _, err := match.Start(context.Background(), b.ID); err != nil {
		t.Fatalf("bob start: %v", err)
	}
}

func TestMatchingRemovesUnreadyQueueEntry(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	alice := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	deleted := testutil.CreateUser(t, repos, "ghost", "ghost@test.local", "password123")
	charlie := testutil.CreateUser(t, repos, "charlie", "charlie@test.local", "password123")
	dave := testutil.CreateUser(t, repos, "dave", "dave@test.local", "password123")

	if _, err := match.Start(context.Background(), alice.ID); err != nil {
		t.Fatalf("alice start: %v", err)
	}
	now := time.Now().UTC()
	if err := repos.MatchingQueue.Insert(context.Background(), &model.MatchingQueueEntry{
		ID: uuid.New(), UserID: deleted.ID, Status: model.MatchingQueueWaiting, CreatedAt: now,
	}); err != nil {
		t.Fatalf("insert ghost: %v", err)
	}
	deleted.DeletedAt = &now
	if err := repos.User.Update(context.Background(), deleted); err != nil {
		t.Fatalf("delete ghost: %v", err)
	}

	if _, err := match.Start(context.Background(), charlie.ID); err != nil {
		t.Fatalf("charlie start: %v", err)
	}
	if _, err := match.Start(context.Background(), dave.ID); err != nil {
		t.Fatalf("dave start: %v", err)
	}

	status, err := match.Status(context.Background(), alice.ID)
	if err != nil || status.Status != "matched" {
		t.Fatalf("alice should be matched: %+v err=%v", status, err)
	}
}

func TestMatchingRemovesMasterFromQueue(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	alice := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	bob := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	if _, err := match.Start(context.Background(), alice.ID); err != nil {
		t.Fatalf("alice start: %v", err)
	}
	now := time.Now().UTC()
	masterEntryID := uuid.New()
	_ = masterEntryID
	if err := repos.MatchingQueue.Insert(context.Background(), &model.MatchingQueueEntry{
		ID: masterEntryID, UserID: master.ID, Status: model.MatchingQueueWaiting, CreatedAt: now,
	}); err != nil {
		t.Fatalf("insert master: %v", err)
	}

	if _, err := match.Start(context.Background(), bob.ID); err != nil {
		t.Fatalf("bob start: %v", err)
	}

	entry, err := repos.MatchingQueue.FindByUserID(context.Background(), master.ID)
	if err != nil {
		t.Fatalf("find master entry: %v", err)
	}
	if entry != nil {
		t.Fatalf("master queue entry should be removed: %+v", entry)
	}

	aliceStatus, err := match.Status(context.Background(), alice.ID)
	if err != nil || aliceStatus.Status != "waiting" {
		t.Fatalf("alice still waiting after master removed: %+v err=%v", aliceStatus, err)
	}
	bobStatus, err := match.Status(context.Background(), bob.ID)
	if err != nil || bobStatus.Status != "waiting" {
		t.Fatalf("bob waiting: %+v err=%v", bobStatus, err)
	}
}

func TestMatchingCustomNow(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	match.Now = func() time.Time { return time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC) }
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	if _, err := match.Start(context.Background(), user.ID); err != nil {
		t.Fatalf("start: %v", err)
	}
}
