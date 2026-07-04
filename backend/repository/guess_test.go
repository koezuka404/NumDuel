package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/numduel/numduel/model"
)

func TestGuessRepoCRUD(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	p1 := createUser(t, repos, "alice", "alice@test.local")
	p2 := createUser(t, repos, "bob", "bob@test.local")
	game := newGame(p1.ID, p2.ID, model.GameStatusInProgress)
	if err := repos.Game.Create(ctx, game); err != nil {
		t.Fatalf("create game: %v", err)
	}

	g1 := newGuess(game.ID, p1.ID, 1)
	g2 := newGuess(game.ID, p2.ID, 1)
	if err := repos.Guess.Create(ctx, g1); err != nil {
		t.Fatalf("create guess1: %v", err)
	}
	if err := repos.Guess.Create(ctx, g2); err != nil {
		t.Fatalf("create guess2: %v", err)
	}

	rows, err := repos.Guess.ListByGameAndPlayer(ctx, game.ID, p1.ID)
	if err != nil || len(rows) != 1 || rows[0].PlayerID != p1.ID {
		t.Fatalf("list by player: %+v err=%v", rows, err)
	}

	count, err := repos.Guess.CountByGameExcludingPlayer(ctx, game.ID, p1.ID)
	if err != nil || count != 1 {
		t.Fatalf("count excluding: %d err=%v", count, err)
	}

	since := time.Now().UTC().Add(-time.Minute)
	updated, err := repos.Guess.FindUpdatedSince(ctx, since)
	if err != nil || len(updated) < 2 {
		t.Fatalf("updated since: %d err=%v", len(updated), err)
	}
}
