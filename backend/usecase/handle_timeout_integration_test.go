package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestHandleTimeoutAutoGuess(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	turns := &memTurnStore{}
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Turns = turns

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	game, _ := repos.Game.FindByID(context.Background(), gameID)
	now := time.Now().UTC()
	turns.turns[gameID] = &usecase.TurnInfo{
		Turn: game.CurrentTurn, PlayerID: a.ID, ExpiresAt: now.Add(-time.Second),
	}

	if err := gameUC.HandleTimeout(context.Background(), gameID, a.ID); err != nil {
		t.Fatalf("handle timeout: %v", err)
	}

	guesses, err := repos.Guess.ListByGameAndPlayer(context.Background(), gameID, a.ID)
	if err != nil || len(guesses) == 0 {
		t.Fatalf("auto guess expected: %+v err=%v", guesses, err)
	}
	if !guesses[0].IsAuto {
		t.Fatalf("expected auto guess, got %+v", guesses[0])
	}

	logs, _, err := repos.ActivityLog.Search(context.Background(), "timeout", &a.ID, nil, nil, 1, 10)
	if err != nil || len(logs) == 0 {
		t.Fatalf("timeout activity log: %+v err=%v", logs, err)
	}
}

func TestHandleTimeoutSkipsWhenNotExpired(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	turns := &memTurnStore{}
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Turns = turns

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	game, _ := repos.Game.FindByID(context.Background(), gameID)
	now := time.Now().UTC()
	turns.turns[gameID] = &usecase.TurnInfo{
		Turn: game.CurrentTurn, PlayerID: a.ID, ExpiresAt: now.Add(time.Minute),
	}

	if err := gameUC.HandleTimeout(context.Background(), gameID, a.ID); err != nil {
		t.Fatalf("handle timeout: %v", err)
	}
	guesses, _ := repos.Guess.ListByGameAndPlayer(context.Background(), gameID, a.ID)
	if len(guesses) != 0 {
		t.Fatalf("expected no guess when turn not expired")
	}
}

func TestHandleTimeoutSkipsWrongPlayer(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	turns := &memTurnStore{}
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Turns = turns

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	game, _ := repos.Game.FindByID(context.Background(), gameID)
	now := time.Now().UTC()
	turns.turns[gameID] = &usecase.TurnInfo{
		Turn: game.CurrentTurn, PlayerID: a.ID, ExpiresAt: now.Add(-time.Second),
	}

	if err := gameUC.HandleTimeout(context.Background(), gameID, b.ID); err != nil {
		t.Fatalf("handle timeout: %v", err)
	}
}

func TestHandleTimeoutNilTurnStore(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Turns = nil

	if err := gameUC.HandleTimeout(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("nil turn store: %v", err)
	}
}

func TestHandleTimeoutLockNotAcquired(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	turns := &memTurnStore{}
	locks := newMemLockStore()
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Turns = turns
	gameUC.Locks = locks

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	game, _ := repos.Game.FindByID(context.Background(), gameID)
	now := time.Now().UTC()
	turns.turns[gameID] = &usecase.TurnInfo{
		Turn: game.CurrentTurn, PlayerID: a.ID, ExpiresAt: now.Add(-time.Second),
	}
	locks.locked["game:"+gameID.String()+":player:"+a.ID.String()+":guess_lock"] = true

	if err := gameUC.HandleTimeout(context.Background(), gameID, a.ID); err != nil {
		t.Fatalf("lock held: %v", err)
	}
}
