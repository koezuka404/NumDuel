package usecase_test

import (
	"context"
	"testing"

	"github.com/numduel/numduel/testutil"
)

func TestSyncGameStateNotifies(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	notifier := &captureNotifier{}
	gameUC := testutil.NewGameUCWithNotifier(t, repos, notifier)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	if _, err := gameUC.SyncGameState(context.Background(), a.ID, gameID); err != nil {
		t.Fatalf("sync: %v", err)
	}
	found := false
	for _, call := range notifier.calls {
		if call.EventType == "GAME_STATE_SYNC" && call.UserID == a.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected GAME_STATE_SYNC notification: %+v", notifier.calls)
	}
}

func TestSubmitGuessNotifiesTurnChanged(t *testing.T) {
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

	if err := gameUC.SubmitGuess(context.Background(), a.ID, gameID, "9012", false); err != nil {
		t.Fatalf("guess: %v", err)
	}
	found := false
	for _, call := range notifier.calls {
		if call.EventType == "TURN_CHANGED" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected TURN_CHANGED notification: %+v", notifier.calls)
	}
}

func TestSubmitGuessNotifiesGuessResult(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	notifier := &captureNotifier{}
	gameUC := testutil.NewGameUCWithNotifier(t, repos, notifier)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	if err := gameUC.SubmitGuess(context.Background(), a.ID, gameID, "9012", false); err != nil {
		t.Fatalf("guess: %v", err)
	}
	found := false
	for _, call := range notifier.calls {
		if call.EventType == "GUESS_RESULT" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected GUESS_RESULT notification: %+v", notifier.calls)
	}
}
