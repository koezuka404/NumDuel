package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
)

func openMigratedDBFailSecondQuery(t *testing.T, models ...any) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return registerFailSecondQueryHook(t, gdb, models...)
}

func openPostgresDBFailSecondQuery(t *testing.T, models ...any) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(postgres.Open(testPrimaryURL(t)), &gorm.Config{})
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	return registerFailSecondQueryHook(t, gdb, models...)
}

func registerFailSecondQueryHook(t *testing.T, gdb *gorm.DB, models ...any) *gorm.DB {
	t.Helper()
	if err := gdb.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	queryN := 0
	hookName := "fail_second_query_" + t.Name()
	if err := gdb.Callback().Query().Before("gorm:query").Register(hookName, func(db *gorm.DB) {
		queryN++
		if queryN >= 2 {
			_ = db.AddError(errors.New("injected query error"))
		}
	}); err != nil {
		t.Fatalf("register callback: %v", err)
	}
	return gdb
}

func openUnmigratedDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return gdb
}

func TestGameRepoListErrors(t *testing.T) {
	repo := repository.NewGameRepo(openUnmigratedDB(t))
	ctx := context.Background()
	uid := uuid.New()
	before := time.Now().UTC()

	if _, err := repo.ListByPlayerID(ctx, uid); err == nil {
		t.Fatal("expected ListByPlayerID error")
	}
	if _, err := repo.ListByStatus(ctx, model.GameStatusInProgress); err == nil {
		t.Fatal("expected ListByStatus error")
	}
	if _, err := repo.ListByStatusCreatedBefore(ctx, model.GameStatusFinished, before); err == nil {
		t.Fatal("expected ListByStatusCreatedBefore error")
	}
	if _, err := repo.FindUpdatedSince(ctx, before); err == nil {
		t.Fatal("expected FindUpdatedSince error")
	}
}

func TestUserRepoListErrors(t *testing.T) {
	repo := repository.NewUserRepo(openUnmigratedDB(t))
	ctx := context.Background()
	before := time.Now().UTC()

	if _, err := repo.ListAll(ctx); err == nil {
		t.Fatal("expected ListAll error")
	}
	if _, _, err := repo.List(ctx, 1, 10); err == nil {
		t.Fatal("expected List error")
	}
	if _, _, err := repo.Search(ctx, "ali", 1, 10); err == nil {
		t.Fatal("expected Search error")
	}
	if _, err := repo.FindUpdatedSince(ctx, before); err == nil {
		t.Fatal("expected FindUpdatedSince error")
	}
	if _, err := repo.ListInactiveSince(ctx, before); err == nil {
		t.Fatal("expected ListInactiveSince error")
	}
}

func TestLoginLogRepoListErrors(t *testing.T) {
	repo := repository.NewLoginLogRepo(openUnmigratedDB(t))
	ctx := context.Background()

	if _, _, err := repo.ListByUserID(ctx, uuid.New(), 1, 10); err == nil {
		t.Fatal("expected ListByUserID error")
	}
}

func TestWSConnectionLogRepoListErrors(t *testing.T) {
	repo := repository.NewWSConnectionLogRepo(openUnmigratedDB(t))
	ctx := context.Background()

	if _, _, err := repo.ListByUserID(ctx, uuid.New(), 1, 10); err == nil {
		t.Fatal("expected ListByUserID error")
	}
}

func TestMatchHistoryRepoListErrors(t *testing.T) {
	repo := repository.NewMatchHistoryRepo(openUnmigratedDB(t))
	ctx := context.Background()

	if _, _, err := repo.ListByUserID(ctx, uuid.New(), 1, 10); err == nil {
		t.Fatal("expected ListByUserID error")
	}
}

func TestActivityLogRepoSearchErrors(t *testing.T) {
	repo := repository.NewActivityLogRepo(openUnmigratedDB(t))
	ctx := context.Background()

	if _, _, err := repo.Search(ctx, "login", nil, nil, nil, 1, 10); err == nil {
		t.Fatal("expected Search error")
	}
}

func TestActivityLogRepoSearchFindError(t *testing.T) {
	repo := repository.NewActivityLogRepo(openMigratedDBFailSecondQuery(t, &model.ActivityLog{}))
	ctx := context.Background()

	if _, _, err := repo.Search(ctx, "", nil, nil, nil, 1, 10); err == nil {
		t.Fatal("expected Search find error")
	}
}

func TestLoginLogRepoListByUserIDFindError(t *testing.T) {
	repo := repository.NewLoginLogRepo(openMigratedDBFailSecondQuery(t, &model.LoginLog{}))
	ctx := context.Background()

	if _, _, err := repo.ListByUserID(ctx, uuid.New(), 1, 10); err == nil {
		t.Fatal("expected ListByUserID find error")
	}
}

func TestWSConnectionLogRepoListByUserIDFindError(t *testing.T) {
	repo := repository.NewWSConnectionLogRepo(openMigratedDBFailSecondQuery(t, &model.WSConnectionLog{}))
	ctx := context.Background()

	if _, _, err := repo.ListByUserID(ctx, uuid.New(), 1, 10); err == nil {
		t.Fatal("expected ListByUserID find error")
	}
}

func TestMatchHistoryRepoListByUserIDFindError(t *testing.T) {
	repo := repository.NewMatchHistoryRepo(openMigratedDBFailSecondQuery(t, &model.MatchHistory{}))
	ctx := context.Background()

	if _, _, err := repo.ListByUserID(ctx, uuid.New(), 1, 10); err == nil {
		t.Fatal("expected ListByUserID find error")
	}
}

func TestUserRepoListFindError(t *testing.T) {
	repo := repository.NewUserRepo(openMigratedDBFailSecondQuery(t, &model.User{}))
	ctx := context.Background()

	if _, _, err := repo.List(ctx, 1, 10); err == nil {
		t.Fatal("expected List find error")
	}
}

func TestUserRepoSearchFindError(t *testing.T) {
	repo := repository.NewUserRepo(openMigratedDBFailSecondQuery(t, &model.User{}))
	ctx := context.Background()

	if _, _, err := repo.Search(ctx, "ali", 1, 10); err == nil {
		t.Fatal("expected Search find error")
	}
}

func TestRankingRepoReplaceAllError(t *testing.T) {
	repo := repository.NewRankingRepo(openUnmigratedDB(t))
	ctx := context.Background()
	now := time.Now().UTC()

	err := repo.ReplaceAll(ctx, []model.Ranking{{
		UserID: uuid.New(), Rank: 1, Username: "alice", WinCount: 1, UpdatedAt: now,
	}})
	if err == nil {
		t.Fatal("expected ReplaceAll error without rankings table")
	}
}

func TestRepoErrorsAfterClose(t *testing.T) {
	gdb, repos := testutil.OpenSQLiteDB(t)
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	ctx := context.Background()
	uid := uuid.New()
	now := time.Now().UTC()

	if _, _, err := repos.User.List(ctx, 1, 10); err == nil {
		t.Fatal("expected closed-db List error")
	}
	if _, _, err := repos.User.Search(ctx, "ali", 1, 10); err == nil {
		t.Fatal("expected closed-db Search error")
	}
	if _, _, err := repos.ActivityLog.Search(ctx, "", nil, nil, nil, 1, 10); err == nil {
		t.Fatal("expected closed-db activity Search error")
	}
	if _, _, err := repos.LoginLog.ListByUserID(ctx, uid, 1, 10); err == nil {
		t.Fatal("expected closed-db login ListByUserID error")
	}
	if _, _, err := repos.WSConnectionLog.ListByUserID(ctx, uid, 1, 10); err == nil {
		t.Fatal("expected closed-db ws ListByUserID error")
	}
	if _, _, err := repos.MatchHistory.ListByUserID(ctx, uid, 1, 10); err == nil {
		t.Fatal("expected closed-db match history ListByUserID error")
	}
	_ = now
}

func TestBackupSyncerPrimaryFindError(t *testing.T) {
	primaryGDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open primary: %v", err)
	}
	_, backupRepos := testutil.OpenSQLiteDB(t)
	syncer := repository.NewBackupSyncer(primaryGDB, backupRepos.DB)
	if _, err := syncer.Sync(context.Background(), nil); err == nil {
		t.Fatal("expected sync error when primary schema is missing")
	}
}
