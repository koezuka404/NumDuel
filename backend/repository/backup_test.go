package repository_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
)

func TestBackupSyncer(t *testing.T) {
	primary, backup := openBackupDBs(t)
	repos := repository.NewRepos(primary)
	user := createUser(t, repos, "alice", "alice@test.local")

	syncer := repository.NewBackupSyncer(primary, backup)
	result, err := syncer.Sync(context.Background(), nil)
	if err != nil {
		t.Fatalf("sync all: %v", err)
	}
	if result.SyncedRows == 0 {
		t.Fatalf("expected synced rows")
	}

	since := time.Now().UTC().Add(-time.Minute)
	result, err = syncer.Sync(context.Background(), &since)
	if err != nil {
		t.Fatalf("sync since: %v", err)
	}
	_ = result

	user.WinCount = 9
	if err := repos.User.Update(context.Background(), user); err != nil {
		t.Fatalf("update user: %v", err)
	}
	result, err = syncer.Sync(context.Background(), &since)
	if err != nil {
		t.Fatalf("sync updated: %v", err)
	}
	if result.SyncedRows == 0 {
		t.Fatalf("expected updated user sync")
	}
}

func TestBackupSyncerAllTables(t *testing.T) {
	primary, backup := openBackupDBs(t)
	repos := repository.NewRepos(primary)
	ctx := context.Background()
	p1 := createUser(t, repos, "alice", "alice@test.local")
	p2 := createUser(t, repos, "bob", "bob@test.local")

	game := newGame(p1.ID, p2.ID, model.GameStatusFinished)
	if err := repos.Game.Create(ctx, game); err != nil {
		t.Fatalf("game: %v", err)
	}
	guess := newGuess(game.ID, p1.ID, 1)
	if err := repos.Guess.Create(ctx, guess); err != nil {
		t.Fatalf("guess: %v", err)
	}
	history := newMatchHistory(game.ID, p1.ID, p2.ID)
	if err := repos.MatchHistory.Create(ctx, history); err != nil {
		t.Fatalf("history: %v", err)
	}
	if err := repos.Ranking.ReplaceAll(ctx, []model.Ranking{{
		UserID: p1.ID, Rank: 1, Username: "alice", WinCount: 1, UpdatedAt: time.Now().UTC(),
	}}); err != nil {
		t.Fatalf("ranking: %v", err)
	}
	if err := repos.ActivityLog.Create(ctx, newActivityLog("sync", &p1.ID)); err != nil {
		t.Fatalf("activity: %v", err)
	}
	if err := repos.LoginLog.Create(ctx, newLoginLog(p1.ID, model.LoginActionLogin)); err != nil {
		t.Fatalf("login: %v", err)
	}

	syncer := repository.NewBackupSyncer(primary, backup)
	result, err := syncer.Sync(ctx, nil)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if result.SyncedRows == 0 {
		t.Fatal("expected rows synced")
	}
}

func openBackupDBs(t *testing.T) (*gorm.DB, *gorm.DB) {
	t.Helper()
	_, repos := testutil.OpenSQLiteDB(t)
	_, backupRepos := testutil.OpenSQLiteDB(t)
	return repos.DB, backupRepos.DB
}

func TestBackupSyncerBackupFailure(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	createUser(t, repos, "alice", "alice@test.local")

	backupGDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}

	syncer := repository.NewBackupSyncer(repos.DB, backupGDB)
	if _, err := syncer.Sync(context.Background(), nil); err == nil {
		t.Fatal("expected sync error when backup schema is missing")
	}
}
