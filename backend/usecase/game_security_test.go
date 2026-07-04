package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

// セキュリティ: IDOR・秘密数字漏洩防止
func TestGetGameStateForbiddenForNonParticipant(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	c := testutil.CreateUser(t, repos, "carol", "carol@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	_, err := gameUC.GetGameState(context.Background(), c.ID, gameID)
	if !errors.Is(err, usecase.ErrForbidden) {
		t.Fatalf("non-participant: %v", err)
	}
}

func TestGameStateDoesNotExposeSecretHashes(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	game, err := repos.Game.FindByID(context.Background(), gameID)
	if err != nil || game == nil {
		t.Fatalf("game: %v", err)
	}

	state, err := gameUC.GetGameState(context.Background(), a.ID, gameID)
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	snapshot := state.GameID.String() + state.CurrentTurnPlayerID + string(state.Status)
	for _, secret := range []string{game.Player1Secret, game.Player2Secret} {
		if secret != "" && strings.Contains(snapshot, secret) {
			t.Fatalf("secret hash leaked in game state output")
		}
	}
}

func TestSetSecretForbiddenForNonParticipant(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	c := testutil.CreateUser(t, repos, "carol", "carol@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	err := gameUC.SetSecretNumber(context.Background(), c.ID, gameID, "9012")
	if !errors.Is(err, usecase.ErrForbidden) {
		t.Fatalf("non-participant set secret: %v", err)
	}
}

func TestDeletedUserCannotSetSecret(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	now := time.Now().UTC()
	a.DeletedAt = &now
	if err := repos.User.Update(context.Background(), a); err != nil {
		t.Fatalf("mark deleted: %v", err)
	}

	err := gameUC.SetSecretNumber(context.Background(), a.ID, gameID, "1234")
	if !errors.Is(err, usecase.ErrUnauthorized) {
		t.Fatalf("deleted user set secret: %v", err)
	}
	_ = b
}
