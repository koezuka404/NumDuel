package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/numduel/numduel/model"
)

func TestMatchHistoryRepo(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()
	winner := createUser(t, repos, "alice", "alice@test.local")
	loser := createUser(t, repos, "bob", "bob@test.local")
	game := newGame(winner.ID, loser.ID, model.GameStatusFinished)

	history := newMatchHistory(game.ID, winner.ID, loser.ID)
	if err := repos.MatchHistory.Create(ctx, history); err != nil {
		t.Fatalf("create: %v", err)
	}

	rows, total, err := repos.MatchHistory.ListByUserID(ctx, winner.ID, 1, 10)
	if err != nil || total != 1 || len(rows) != 1 {
		t.Fatalf("list winner: total=%d len=%d err=%v", total, len(rows), err)
	}

	rows, total, err = repos.MatchHistory.ListByUserID(ctx, loser.ID, 0, 0)
	if err != nil || total != 1 {
		t.Fatalf("list loser: total=%d err=%v", total, err)
	}

	since := time.Now().UTC().Add(-time.Minute)
	updated, err := repos.MatchHistory.FindUpdatedSince(ctx, since)
	if err != nil || len(updated) != 1 {
		t.Fatalf("updated since: %d err=%v", len(updated), err)
	}
}
