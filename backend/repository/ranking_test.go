package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestRankingRepo(t *testing.T) {
	repos := openRepos(t)
	ctx := context.Background()

	if err := repos.Ranking.ReplaceAll(ctx, nil); err != nil {
		t.Fatalf("replace empty: %v", err)
	}
	all, err := repos.Ranking.ListAll(ctx)
	if err != nil || len(all) != 0 {
		t.Fatalf("list empty: %d err=%v", len(all), err)
	}

	now := time.Now().UTC()
	rows := []model.Ranking{
		{UserID: uuid.New(), Rank: 1, Username: "alice", WinCount: 5, UpdatedAt: now},
		{UserID: uuid.New(), Rank: 2, Username: "bob", WinCount: 3, UpdatedAt: now},
	}
	if err := repos.Ranking.ReplaceAll(ctx, rows); err != nil {
		t.Fatalf("replace rows: %v", err)
	}

	all, err = repos.Ranking.ListAll(ctx)
	if err != nil || len(all) != 2 || all[0].Rank != 1 {
		t.Fatalf("list all: %+v err=%v", all, err)
	}

	since := now.Add(-time.Minute)
	updated, err := repos.Ranking.FindUpdatedSince(ctx, since)
	if err != nil || len(updated) != 2 {
		t.Fatalf("updated since: %d err=%v", len(updated), err)
	}
}
