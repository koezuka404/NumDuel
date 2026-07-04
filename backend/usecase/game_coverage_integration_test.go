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

func TestHandleTimeoutGameFinished(t *testing.T) {
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
	game.Status = model.GameStatusFinished
	if err := repos.Game.Update(context.Background(), game); err != nil {
		t.Fatalf("finish game: %v", err)
	}

	now := time.Now().UTC()
	turns.turns[gameID] = &usecase.TurnInfo{
		Turn: game.CurrentTurn, PlayerID: a.ID, ExpiresAt: now.Add(-time.Second),
	}

	if err := gameUC.HandleTimeout(context.Background(), gameID, a.ID); err != nil {
		t.Fatalf("handle timeout finished game: %v", err)
	}
}

func TestHandleTimeoutTurnMismatch(t *testing.T) {
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
		Turn: game.CurrentTurn + 1, PlayerID: a.ID, ExpiresAt: now.Add(-time.Second),
	}

	if err := gameUC.HandleTimeout(context.Background(), gameID, a.ID); err != nil {
		t.Fatalf("turn mismatch: %v", err)
	}
}

func TestHandleTimeoutNilRandom(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	turns := &memTurnStore{}
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Turns = turns
	gameUC.Random = nil

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	game, _ := repos.Game.FindByID(context.Background(), gameID)
	now := time.Now().UTC()
	turns.turns[gameID] = &usecase.TurnInfo{
		Turn: game.CurrentTurn, PlayerID: a.ID, ExpiresAt: now.Add(-time.Second),
	}

	err := gameUC.HandleTimeout(context.Background(), gameID, a.ID)
	if err == nil {
		t.Fatalf("expected error when random is nil")
	}
}

func TestHandleTimeoutNilTurnAfterValidation(t *testing.T) {
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
	delete(turns.turns, gameID)

	if err := gameUC.HandleTimeout(context.Background(), gameID, a.ID); err != nil {
		t.Fatalf("nil turn after validation: %v", err)
	}
}

func TestCancelBySecretTimeoutNotExpired(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	if err := gameUC.CancelBySecretTimeout(context.Background(), gameID); err != nil {
		t.Fatalf("cancel not expired: %v", err)
	}
	game, _ := repos.Game.FindByID(context.Background(), gameID)
	if game.Status != model.GameStatusWaitingSecret {
		t.Fatalf("game should remain waiting: %s", game.Status)
	}
}

func TestCancelBySecretTimeoutWithNotifier(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	notifier := &captureNotifier{}
	turns := &memTurnStore{}
	gameUC := testutil.NewGameUCWithNotifier(t, repos, notifier)
	gameUC.Turns = turns

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	game, _ := repos.Game.FindByID(context.Background(), gameID)
	game.CreatedAt = time.Now().UTC().Add(-2 * time.Minute)
	if err := repos.Game.Update(context.Background(), game); err != nil {
		t.Fatalf("update created_at: %v", err)
	}

	if err := gameUC.CancelBySecretTimeout(context.Background(), gameID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if len(notifier.calls) < 2 {
		t.Fatalf("expected GAME_OVER notifications: %+v", notifier.calls)
	}
}

func TestCancelBySecretTimeoutMissingGame(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	gameUC := testutil.NewGameUC(t, repos)

	if err := gameUC.CancelBySecretTimeout(context.Background(), uuid.New()); err != nil {
		t.Fatalf("missing game: %v", err)
	}
}

func TestRecoverActiveGamesSkipsValidTurn(t *testing.T) {
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
	expires := time.Now().UTC().Add(time.Minute)
	turns.turns[gameID] = &usecase.TurnInfo{
		Turn: game.CurrentTurn, PlayerID: a.ID, ExpiresAt: expires,
	}

	if err := gameUC.RecoverActiveGames(context.Background()); err != nil {
		t.Fatalf("recover: %v", err)
	}
	if !turns.turns[gameID].ExpiresAt.Equal(expires) {
		t.Fatalf("valid turn should not be replaced: %+v", turns.turns[gameID])
	}
}

func TestRecoverActiveGamesWithNotifier(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	notifier := &captureNotifier{}
	turns := &memTurnStore{}
	gameUC := testutil.NewGameUCWithNotifier(t, repos, notifier)
	gameUC.Turns = turns

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	game, _ := repos.Game.FindByID(context.Background(), gameID)
	turns.turns[gameID] = &usecase.TurnInfo{
		Turn: game.CurrentTurn, PlayerID: a.ID, ExpiresAt: time.Now().UTC().Add(-time.Second),
	}

	if err := gameUC.RecoverActiveGames(context.Background()); err != nil {
		t.Fatalf("recover: %v", err)
	}
	found := false
	for _, call := range notifier.calls {
		if call.EventType == "TURN_CHANGED" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected TURN_CHANGED on recover: %+v", notifier.calls)
	}
}

func TestSetSecretLockHeld(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	locks := newMemLockStore()
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Locks = locks

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	lockKey := "game:" + gameID.String() + ":player:" + a.ID.String() + ":secret_lock"
	locks.locked[lockKey] = true

	err := gameUC.SetSecretNumber(context.Background(), a.ID, gameID, "1234")
	if !errors.Is(err, usecase.ErrRateLimitExceeded) {
		t.Fatalf("secret lock held: %v", err)
	}
}

func TestSetSecretGameNotFound(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	gameUC := testutil.NewGameUC(t, repos)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	err := gameUC.SetSecretNumber(context.Background(), user.ID, uuid.New(), "1234")
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Fatalf("game not found: %v", err)
	}
}

func TestSubmitGuessWinNotifiesGameOver(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	notifier := &captureNotifier{}
	gameUC := testutil.NewGameUCWithNotifier(t, repos, notifier)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	if err := gameUC.SubmitGuess(context.Background(), a.ID, gameID, "9012", false); err != nil {
		t.Fatalf("miss: %v", err)
	}
	if err := gameUC.SubmitGuess(context.Background(), b.ID, gameID, "1234", false); err != nil {
		t.Fatalf("win: %v", err)
	}

	found := false
	for _, call := range notifier.calls {
		if call.EventType == "GAME_OVER" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected GAME_OVER: %+v", notifier.calls)
	}
}

func TestGameCustomNow(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	gameUC := testutil.NewGameUC(t, repos)
	fixed := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	gameUC.Now = func() time.Time { return fixed }
	_ = fixed
}

func TestSubmitGuessInvalidDigits(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	err := gameUC.SubmitGuess(context.Background(), a.ID, gameID, "12a4", false)
	if !errors.Is(err, model.ErrBadDigit) {
		t.Fatalf("invalid digits: %v", err)
	}
}

func TestSubmitGuessLockHeld(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	locks := newMemLockStore()
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Locks = locks
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	lockKey := "game:" + gameID.String() + ":player:" + a.ID.String() + ":guess_lock"
	locks.locked[lockKey] = true

	err := gameUC.SubmitGuess(context.Background(), a.ID, gameID, "9012", false)
	if !errors.Is(err, usecase.ErrRateLimitExceeded) {
		t.Fatalf("guess lock held: %v", err)
	}
}

func TestSubmitGuessMasterForbidden(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	err := gameUC.SubmitGuess(context.Background(), master.ID, gameID, "9012", false)
	if !errors.Is(err, usecase.ErrForbidden) {
		t.Fatalf("master guess forbidden: %v", err)
	}
}

func TestSetSecretDuplicate(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	if err := gameUC.SetSecretNumber(context.Background(), a.ID, gameID, "1234"); err != nil {
		t.Fatalf("first secret: %v", err)
	}
	err := gameUC.SetSecretNumber(context.Background(), a.ID, gameID, "5678")
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("duplicate secret: %v", err)
	}
	_ = b
}

func TestSetSecretGameAlreadyStarted(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)
	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	err := gameUC.SetSecretNumber(context.Background(), a.ID, gameID, "9012")
	if !errors.Is(err, usecase.ErrGameAlreadyStarted) {
		t.Fatalf("already started: %v", err)
	}
}

func TestHandleTimeoutLockError(t *testing.T) {
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
	lockKey := "game:" + gameID.String() + ":player:" + a.ID.String() + ":guess_lock"
	locks.errKey = lockKey

	err := gameUC.HandleTimeout(context.Background(), gameID, a.ID)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("lock error: %v", err)
	}
}

func TestRecoverActiveGamesNilTurnStore(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)
	gameUC.Turns = nil

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	if err := gameUC.RecoverActiveGames(context.Background()); err != nil {
		t.Fatalf("recover nil turns: %v", err)
	}
}
