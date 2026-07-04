package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

type memTurnStore struct {
	turns map[uuid.UUID]*usecase.TurnInfo
}

func (m *memTurnStore) SetTurn(_ context.Context, gameID uuid.UUID, turn int, playerID uuid.UUID, _, expiresAt time.Time) error {
	if m.turns == nil {
		m.turns = make(map[uuid.UUID]*usecase.TurnInfo)
	}
	m.turns[gameID] = &usecase.TurnInfo{Turn: turn, PlayerID: playerID, ExpiresAt: expiresAt}
	return nil
}

func (m *memTurnStore) GetTurn(_ context.Context, gameID uuid.UUID) (*usecase.TurnInfo, error) {
	if m.turns == nil {
		return nil, nil
	}
	return m.turns[gameID], nil
}

func (m *memTurnStore) RemainingSeconds(_ context.Context, gameID uuid.UUID, now time.Time) (int, error) {
	info := m.turns[gameID]
	if info == nil {
		return 0, nil
	}
	sec := int(info.ExpiresAt.Sub(now).Seconds())
	if sec < 0 {
		return 0, nil
	}
	return sec, nil
}

func (m *memTurnStore) DeleteTurn(_ context.Context, gameID uuid.UUID) error {
	delete(m.turns, gameID)
	return nil
}

// §18.5.3 ゲーム復旧・タイムアウト
func TestCancelBySecretTimeout(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	game, err := repos.Game.FindByID(context.Background(), gameID)
	if err != nil || game == nil {
		t.Fatalf("game: %v", err)
	}
	game.CreatedAt = time.Now().UTC().Add(-2 * time.Minute)
	if err := repos.Game.Update(context.Background(), game); err != nil {
		t.Fatalf("update created_at: %v", err)
	}

	if err := gameUC.CancelBySecretTimeout(context.Background(), gameID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	game, _ = repos.Game.FindByID(context.Background(), gameID)
	if game.Status != model.GameStatusFinished || game.FinishedAt == nil {
		t.Fatalf("cancelled game: %+v", game)
	}
}

func TestRecoverActiveGamesSecretTimeout(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	match := testutil.NewMatchingUC(repos)
	gameUC := testutil.NewGameUC(t, repos)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)

	game, _ := repos.Game.FindByID(context.Background(), gameID)
	game.CreatedAt = time.Now().UTC().Add(-2 * time.Minute)
	if err := repos.Game.Update(context.Background(), game); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := gameUC.RecoverActiveGames(context.Background()); err != nil {
		t.Fatalf("recover: %v", err)
	}
	game, _ = repos.Game.FindByID(context.Background(), gameID)
	if game.Status != model.GameStatusFinished {
		t.Fatalf("recovered game: %+v", game)
	}
}

func TestRecoverActiveGamesRestoresTurn(t *testing.T) {
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
	turns.turns[gameID] = &usecase.TurnInfo{
		Turn: game.CurrentTurn, PlayerID: a.ID, ExpiresAt: time.Now().UTC().Add(-time.Second),
	}

	if err := gameUC.RecoverActiveGames(context.Background()); err != nil {
		t.Fatalf("recover: %v", err)
	}
	info := turns.turns[gameID]
	if info == nil || !info.ExpiresAt.After(time.Now().UTC()) {
		t.Fatalf("turn not restored: %+v", info)
	}
}
