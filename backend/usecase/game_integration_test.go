package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

// §18.5.3 ゲーム系
func TestSetSecretAndStartGame(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	if err := gameUC.SetSecretNumber(context.Background(), a.ID, gameID, "1234"); err != nil {
		t.Fatalf("alice secret: %v", err)
	}
	game, _ := repos.Game.FindByID(context.Background(), gameID)
	if game.Status != model.GameStatusWaitingSecret {
		t.Fatalf("status after one secret: %s", game.Status)
	}

	if err := gameUC.SetSecretNumber(context.Background(), b.ID, gameID, "5678"); err != nil {
		t.Fatalf("bob secret: %v", err)
	}
	game, _ = repos.Game.FindByID(context.Background(), gameID)
	if game.Status != model.GameStatusInProgress {
		t.Fatalf("status after both secrets: %s", game.Status)
	}
}

func TestSubmitGuessNotYourTurn(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	err := gameUC.SubmitGuess(context.Background(), b.ID, gameID, "9012", false)
	if !errors.Is(err, usecase.ErrNotYourTurn) {
		t.Fatalf("not your turn: %v", err)
	}
}

func TestSubmitGuessWin(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	if err := gameUC.SubmitGuess(context.Background(), a.ID, gameID, "9012", false); err != nil {
		t.Fatalf("alice miss guess: %v", err)
	}
	if err := gameUC.SubmitGuess(context.Background(), b.ID, gameID, "1234", false); err != nil {
		t.Fatalf("bob win guess: %v", err)
	}

	game, err := repos.Game.FindByID(context.Background(), gameID)
	if err != nil || game.Status != model.GameStatusFinished || game.WinnerID == nil || *game.WinnerID != b.ID {
		t.Fatalf("finished game: %+v err=%v", game, err)
	}
	winner, err := repos.User.FindByID(context.Background(), b.ID)
	if err != nil || winner.WinCount != 1 {
		t.Fatalf("winner winCount=%d err=%v", winner.WinCount, err)
	}
}

func TestSyncGameStateForbidden(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	c := testutil.CreateUser(t, repos, "carol", "carol@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	_, err := gameUC.SyncGameState(context.Background(), c.ID, gameID)
	if !errors.Is(err, usecase.ErrForbidden) {
		t.Fatalf("forbidden: %v", err)
	}
}

func matchTwo(t *testing.T, match *usecase.MatchingUseCase, a, b uuid.UUID) uuid.UUID {
	t.Helper()
	if _, err := match.Start(context.Background(), a); err != nil {
		t.Fatalf("start a: %v", err)
	}
	if _, err := match.Start(context.Background(), b); err != nil {
		t.Fatalf("start b: %v", err)
	}
	status, err := match.Status(context.Background(), a)
	if err != nil || status.GameID == nil {
		t.Fatalf("status: %+v err=%v", status, err)
	}
	return *status.GameID
}

func setBothSecrets(t *testing.T, gameUC *usecase.GameUseCase, gameID, a, b uuid.UUID, secretA, secretB string) {
	t.Helper()
	if err := gameUC.SetSecretNumber(context.Background(), a, gameID, secretA); err != nil {
		t.Fatalf("secret a: %v", err)
	}
	if err := gameUC.SetSecretNumber(context.Background(), b, gameID, secretB); err != nil {
		t.Fatalf("secret b: %v", err)
	}
}
