package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestRankingGetTopThree(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	admin := testutil.NewAdminUC(repos, ranking)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	for i, name := range []string{"user1", "user2", "user3", "user4", "user5"} {
		u := testutil.CreateUser(t, repos, name, name+"@test.local", "password123")
		u.WinCount = 5 - i
		if err := repos.User.Update(context.Background(), u); err != nil {
			t.Fatalf("update %s: %v", name, err)
		}
	}
	if err := admin.RebuildRanking(context.Background(), master.ID); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	rows, err := ranking.Get(context.Background())
	if err != nil || len(rows) != 3 || rows[0].Username != "user1" {
		t.Fatalf("top three: %+v err=%v", rows, err)
	}
}

func TestRunScheduledRebuild(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	locks := newMemLockStore()
	ranking := usecase.NewRankingUseCase(repos, locks, 5*time.Second)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	a.WinCount = 3
	b.WinCount = 1
	ctx := context.Background()
	if err := repos.User.Update(ctx, a); err != nil {
		t.Fatalf("update alice: %v", err)
	}
	if err := repos.User.Update(ctx, b); err != nil {
		t.Fatalf("update bob: %v", err)
	}

	if err := ranking.RunScheduledRebuild(ctx); err != nil {
		t.Fatalf("scheduled rebuild: %v", err)
	}
	rows, err := ranking.Get(ctx)
	if err != nil || len(rows) != 2 || rows[0].Username != "alice" {
		t.Fatalf("ranking after rebuild: %+v err=%v", rows, err)
	}
}

func TestRunScheduledRebuildSkipsWhenLockHeld(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	locks := newMemLockStore()
	ranking := usecase.NewRankingUseCase(repos, locks, 5*time.Second)

	locks.locked["admin:00000000-0000-0000-0000-000000000000:ranking_rebuild_lock"] = true

	if err := ranking.RunScheduledRebuild(context.Background()); err != nil {
		t.Fatalf("scheduled rebuild: %v", err)
	}
}
