package usecase_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/numduel/numduel/db"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func openBackupDBs(t *testing.T) (*gorm.DB, *gorm.DB) {
	t.Helper()
	primary, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open primary: %v", err)
	}
	backup, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}
	if err := db.Migrate(primary); err != nil {
		t.Fatalf("migrate primary: %v", err)
	}
	if err := db.Migrate(backup); err != nil {
		t.Fatalf("migrate backup: %v", err)
	}
	return primary, backup
}

func TestBackupRunSyncSuccess(t *testing.T) {
	primary, backupDB := openBackupDBs(t)
	repos := repository.NewRepos(primary)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	_ = user

	statusStore := &memBackupStatusStore{}
	syncer := repository.NewBackupSyncer(primary, backupDB)
	backupUC := usecase.NewBackupUseCase(syncer, statusStore, 3)
	syncTime := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	backupUC.Now = func() time.Time { return syncTime }

	if err := backupUC.RunSync(context.Background()); err != nil {
		t.Fatalf("run sync: %v", err)
	}
	if statusStore.status.Status != "ok" {
		t.Fatalf("status: %+v", statusStore.status)
	}
	if statusStore.status.LastSyncedAt == nil || !statusStore.status.LastSyncedAt.Equal(syncTime) {
		t.Fatalf("last synced: %+v", statusStore.status.LastSyncedAt)
	}
}

func TestBackupRunSyncNilSyncer(t *testing.T) {
	backupUC := usecase.NewBackupUseCase(nil, nil, 0)
	if err := backupUC.RunSync(context.Background()); err != nil {
		t.Fatalf("nil syncer: %v", err)
	}
}

func TestBackupRunSyncPreservesLastSyncedOnError(t *testing.T) {
	primary, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open primary: %v", err)
	}
	if err := db.Migrate(primary); err != nil {
		t.Fatalf("migrate primary: %v", err)
	}
	repos := repository.NewRepos(primary)
	testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	backupDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}

	statusStore := &memBackupStatusStore{}
	prev := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	statusStore.status = usecase.BackupStatus{Status: "ok", LastSyncedAt: &prev}

	syncer := repository.NewBackupSyncer(primary, backupDB)
	backupUC := usecase.NewBackupUseCase(syncer, statusStore, 1)
	if err := backupUC.RunSync(context.Background()); err == nil {
		t.Fatalf("expected sync error without backup schema")
	}
	if statusStore.status.Status != "error" {
		t.Fatalf("status after error: %+v", statusStore.status)
	}
	if statusStore.status.LastSyncedAt == nil || !statusStore.status.LastSyncedAt.Equal(prev) {
		t.Fatalf("last synced not preserved: %+v", statusStore.status.LastSyncedAt)
	}
}

func TestAdminGetBackupStatusNilReader(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))

	out, err := admin.GetBackupStatus(context.Background())
	if err != nil || out.Status != "ok" {
		t.Fatalf("default backup status: %+v err=%v", out, err)
	}
}

func TestAdminGetBackupStatusWithReader(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	synced := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	reader := &memBackupStatusStore{
		status: usecase.BackupStatus{Status: "ok", LastSyncedAt: &synced},
	}
	admin := usecase.NewAdminUseCase(repos, testutil.NewRankingUC(repos), nil, nil, reader, nil, 0)

	out, err := admin.GetBackupStatus(context.Background())
	if err != nil || out.Status != "ok" || out.LastSyncedAt == nil || !out.LastSyncedAt.Equal(synced) {
		t.Fatalf("backup status: %+v err=%v", out, err)
	}
}

func TestAdminSearchActivityLogs(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	now := time.Now().UTC()
	detail, _ := json.Marshal(map[string]string{"action": "test"})
	if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
		ID: uuid.New(), UserID: &user.ID, LogType: "guess", Detail: detail, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create log: %v", err)
	}

	items, total, err := admin.SearchActivityLogs(context.Background(), "guess", &user.ID, nil, nil, 1, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].LogType != "guess" {
		t.Fatalf("search logs: items=%+v total=%d err=%v", items, total, err)
	}
}

func TestAdminListActivityLogTypes(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	now := time.Now().UTC()
	for _, logType := range []string{"guess", "game_over"} {
		if err := repos.ActivityLog.Create(context.Background(), &model.ActivityLog{
			ID: uuid.New(), LogType: logType, Detail: []byte(`{}`), CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			t.Fatalf("create log: %v", err)
		}
	}

	types, err := admin.ListActivityLogTypes(context.Background())
	if err != nil || len(types) < 2 {
		t.Fatalf("log types: %+v err=%v", types, err)
	}
}

func TestAdminListUsers(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	admin := testutil.NewAdminUC(repos, testutil.NewRankingUC(repos))
	testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	users, total, err := admin.ListUsers(context.Background(), 1, 10)
	if err != nil || total < 1 || len(users) == 0 {
		t.Fatalf("list users: users=%+v total=%d err=%v", users, total, err)
	}
}
