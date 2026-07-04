package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestAdminSearchUsersCallsRepo(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	users, err := admin.SearchUsers(context.Background(), "alice")
	if err != nil || len(users) == 0 {
		t.Fatalf("search users: %+v err=%v", users, err)
	}
}

func TestAdminDeleteUserNotFound(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	err := admin.DeleteUser(context.Background(), master.ID, uuid.New())
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Fatalf("not found: %v", err)
	}
}

func TestAdminDeleteUserAlreadyDeleted(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	target := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	now := time.Now().UTC()
	target.DeletedAt = &now
	if err := repos.User.Update(context.Background(), target); err != nil {
		t.Fatalf("delete target: %v", err)
	}

	err := admin.DeleteUser(context.Background(), master.ID, target.ID)
	if !errors.Is(err, usecase.ErrUserAlreadyDeleted) {
		t.Fatalf("already deleted: %v", err)
	}
}

func TestAdminDeleteUserLockHeld(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	locks := newMemLockStore()
	ranking := testutil.NewRankingUC(repos)
	admin := usecase.NewAdminUseCase(repos, ranking, nil, nil, nil, locks, 5*time.Second)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	target := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	lockKey := fmt.Sprintf("admin:%s:user_delete_lock", master.ID)
	locks.locked[lockKey] = true

	err := admin.DeleteUser(context.Background(), master.ID, target.ID)
	if !errors.Is(err, usecase.ErrRateLimitExceeded) {
		t.Fatalf("lock held: %v", err)
	}
}

func TestAdminDeleteUserWithForceLogoutAndWS(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	force := testutil.NewMemForceLogout()
	ws := &memWSSessionStore{}
	locks := newMemLockStore()
	ranking := testutil.NewRankingUC(repos)
	admin := usecase.NewAdminUseCase(repos, ranking, ws, force, nil, locks, 5*time.Second)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	target := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	if err := admin.DeleteUser(context.Background(), master.ID, target.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !ws.wasDeleted(target.ID) {
		t.Fatalf("ws session should be cleared")
	}
	before, err := force.GetForceLogoutBefore(context.Background(), target.ID)
	if err != nil || before.IsZero() {
		t.Fatalf("force logout not set: %v", before)
	}
}

func TestAdminDownloadCSVLockHeld(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	locks := newMemLockStore()
	admin := usecase.NewAdminUseCase(repos, testutil.NewRankingUC(repos), nil, nil, nil, locks, 5*time.Second)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	lockKey := fmt.Sprintf("admin:%s:log_download_lock", master.ID)
	locks.locked[lockKey] = true

	_, err := admin.DownloadActivityLogsCSV(context.Background(), master.ID, "", nil, nil, nil)
	if !errors.Is(err, usecase.ErrRateLimitExceeded) {
		t.Fatalf("csv lock held: %v", err)
	}
}

func TestAdminRebuildRankingNilRanking(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	locks := newMemLockStore()
	admin := usecase.NewAdminUseCase(repos, nil, nil, nil, nil, locks, 5*time.Second)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	err := admin.RebuildRanking(context.Background(), master.ID)
	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Fatalf("nil ranking: %v", err)
	}
}

func TestAdminRebuildRankingLockHeld(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	locks := newMemLockStore()
	ranking := testutil.NewRankingUC(repos)
	admin := usecase.NewAdminUseCase(repos, ranking, nil, nil, nil, locks, 5*time.Second)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")

	lockKey := fmt.Sprintf("admin:%s:ranking_rebuild_lock", master.ID)
	locks.locked[lockKey] = true

	err := admin.RebuildRanking(context.Background(), master.ID)
	if !errors.Is(err, usecase.ErrRateLimitExceeded) {
		t.Fatalf("rebuild lock held: %v", err)
	}
}

func TestAdminNowCustom(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	ranking := testutil.NewRankingUC(repos)
	admin := usecase.NewAdminUseCase(repos, ranking, nil, nil, nil, newMemLockStore(), 5*time.Second)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	fixed := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	admin.Now = func() time.Time { return fixed }

	if err := admin.RebuildRanking(context.Background(), master.ID); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	logs, _, err := repos.ActivityLog.Search(context.Background(), "admin_rebuild_ranking", &master.ID, nil, nil, 1, 10)
	if err != nil || len(logs) == 0 {
		t.Fatalf("rebuild log: %+v err=%v", logs, err)
	}
}

func TestAdminListUsersErrorPath(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	users, total, err := admin.ListUsers(context.Background(), 1, 10)
	if err != nil || total < 0 {
		t.Fatalf("list users: %+v total=%d err=%v", users, total, err)
	}
}

func TestAdminSearchActivityLogsWithFilter(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	from := time.Now().UTC().Add(-time.Hour)
	to := time.Now().UTC().Add(time.Hour)
	now := time.Now().UTC()
	detail, _ := json.Marshal(map[string]string{"k": "v"})
	if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
		ID: uuid.New(), UserID: &user.ID, LogType: "filtered", Detail: detail, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create log: %v", err)
	}

	items, total, err := admin.SearchActivityLogs(context.Background(), "filtered", &user.ID, &from, &to, 1, 10)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("filtered logs: items=%+v total=%d err=%v", items, total, err)
	}
}

func TestAdminGetBackupStatusError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	reader := &memBackupStatusStore{err: context.Canceled}
	admin := usecase.NewAdminUseCase(repos, testutil.NewRankingUC(repos), nil, nil, reader, nil, 0)

	_, err := admin.GetBackupStatus(context.Background())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("backup status error: %v", err)
	}
}

func TestAdminDeleteUserLockAcquireError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	locks := newMemLockStore()
	admin := usecase.NewAdminUseCase(repos, testutil.NewRankingUC(repos), nil, nil, nil, locks, 5*time.Second)
	master := testutil.SeedMaster(t, repos, "admin@test.local", "adminpass123")
	target := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	lockKey := fmt.Sprintf("admin:%s:user_delete_lock", master.ID)
	locks.errKey = lockKey

	err := admin.DeleteUser(context.Background(), master.ID, target.ID)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("lock acquire error: %v", err)
	}
}
