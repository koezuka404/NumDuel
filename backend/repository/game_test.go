package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestGameRepoCRUD(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	p1 := createUser(t, repos, "alice", "alice@test.local")
	p2 := createUser(t, repos, "bob", "bob@test.local")

	game := newGame(p1.ID, p2.ID, model.GameStatusWaitingSecret)
	if err := repos.Game.Create(ctx, game); err != nil {
		t.Fatalf("create: %v", err)
	}

	game.Status = model.GameStatusInProgress
	turn := p1.ID
	game.CurrentTurnPlayerID = &turn
	if err := repos.Game.Update(ctx, game); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repos.Game.FindByID(ctx, game.ID)
	if err != nil || got == nil || got.Status != model.GameStatusInProgress {
		t.Fatalf("find by id: %+v err=%v", got, err)
	}

	got, err = repos.Game.FindByIDForUpdate(ctx, game.ID)
	if err != nil || got == nil || got.ID != game.ID {
		t.Fatalf("find for update: %+v err=%v", got, err)
	}

	missing, err := repos.Game.FindByID(ctx, uuid.New())
	if err != nil || missing != nil {
		t.Fatalf("missing: %+v err=%v", missing, err)
	}
}

func TestGameRepoQueries(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	p1 := createUser(t, repos, "alice", "alice@test.local")
	p2 := createUser(t, repos, "bob", "bob@test.local")

	old := newGame(p1.ID, p2.ID, model.GameStatusFinished)
	old.CreatedAt = time.Now().UTC().Add(-2 * time.Hour)
	old.UpdatedAt = old.CreatedAt
	if err := repos.Game.Create(ctx, old); err != nil {
		t.Fatalf("create old: %v", err)
	}

	active := newGame(p1.ID, p2.ID, model.GameStatusInProgress)
	if err := repos.Game.Create(ctx, active); err != nil {
		t.Fatalf("create active: %v", err)
	}

	byPlayer, err := repos.Game.ListByPlayerID(ctx, p1.ID)
	if err != nil || len(byPlayer) < 2 {
		t.Fatalf("by player: %d err=%v", len(byPlayer), err)
	}

	byStatus, err := repos.Game.ListByStatus(ctx, model.GameStatusInProgress)
	if err != nil || len(byStatus) == 0 {
		t.Fatalf("by status: %d err=%v", len(byStatus), err)
	}

	before := time.Now().UTC().Add(-time.Hour)
	oldGames, err := repos.Game.ListByStatusCreatedBefore(ctx, model.GameStatusFinished, before)
	if err != nil || len(oldGames) == 0 {
		t.Fatalf("created before: %d err=%v", len(oldGames), err)
	}

	since := time.Now().UTC().Add(-time.Minute)
	updated, err := repos.Game.FindUpdatedSince(ctx, since)
	if err != nil || len(updated) == 0 {
		t.Fatalf("updated since: %d err=%v", len(updated), err)
	}
}
