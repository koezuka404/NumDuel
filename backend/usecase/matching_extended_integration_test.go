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

func TestMatchingNotifiesBothPlayers(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	notifier := &captureNotifier{}
	match := usecase.NewMatchingUseCase(repos, notifier)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")

	if _, err := match.Start(context.Background(), a.ID); err != nil {
		t.Fatalf("alice start: %v", err)
	}
	if _, err := match.Start(context.Background(), b.ID); err != nil {
		t.Fatalf("bob start: %v", err)
	}

	if len(notifier.calls) != 2 {
		t.Fatalf("expected 2 MATCHED notifications, got %d", len(notifier.calls))
	}
	for _, call := range notifier.calls {
		if call.EventType != "MATCHED" {
			t.Fatalf("unexpected event: %+v", call)
		}
	}
}

func TestMatchingStatusIdle(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	status, err := match.Status(context.Background(), user.ID)
	if err != nil || status.Status != "idle" || status.GameID != nil {
		t.Fatalf("idle status: %+v err=%v", status, err)
	}
}

func TestMatchingStatusWaiting(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	if _, err := match.Start(context.Background(), user.ID); err != nil {
		t.Fatalf("start: %v", err)
	}
	status, err := match.Status(context.Background(), user.ID)
	if err != nil || status.Status != "waiting" {
		t.Fatalf("waiting status: %+v err=%v", status, err)
	}
}

func TestMatchingRemovesDeletedPlayerFromQueue(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")

	if _, err := match.Start(context.Background(), a.ID); err != nil {
		t.Fatalf("alice start: %v", err)
	}
	now := time.Now().UTC()
	b.DeletedAt = &now
	if err := repos.User.Update(context.Background(), b); err != nil {
		t.Fatalf("delete bob: %v", err)
	}
	if _, err := match.Start(context.Background(), b.ID); err == nil {
		t.Fatalf("deleted user should not match")
	}
}

func TestMatchingStartUnauthorized(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)

	_, err := match.Start(context.Background(), uuid.New())
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("missing user: %v", err)
	}
}

func TestMatchingStartActiveGame(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	matchTwo(t, match, a.ID, b.ID)

	_, err := match.Start(context.Background(), a.ID)
	if !errors.Is(err, usecase.ErrUserInActiveGame) {
		t.Fatalf("active game: %v", err)
	}
}

func TestFindActiveGameFinishedIgnored(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	now := time.Now().UTC()
	game := &model.Game{
		ID: uuid.New(), Status: model.GameStatusFinished,
		Player1ID: a.ID, Player2ID: b.ID, CurrentTurn: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.Game.Create(context.Background(), game); err != nil {
		t.Fatalf("create game: %v", err)
	}
	match := testutil.NewMatchingUC(repos)
	status, err := match.Status(context.Background(), a.ID)
	if err != nil || status.Status != "idle" {
		t.Fatalf("finished game should not count as active: %+v err=%v", status, err)
	}
}
