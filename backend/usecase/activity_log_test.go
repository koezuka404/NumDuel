package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/numduel/numduel/db"
	"github.com/numduel/numduel/repository"
)

func openTestRepos(t *testing.T) repository.Repos {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Migrate(gdb); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return repository.NewRepos(gdb)
}

func TestRecordActivityLogMarshalError(t *testing.T) {
	repos := openTestRepos(t)
	err := recordActivityLog(context.Background(), repos, nil, "bad", map[string]any{
		"channel": make(chan int),
	}, time.Now().UTC())
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestRecordActivityLogSuccess(t *testing.T) {
	repos := openTestRepos(t)
	now := time.Now().UTC()
	if err := recordRecoverActivityLog(context.Background(), repos, uuid.New(), now); err != nil {
		t.Fatalf("record recover log: %v", err)
	}
	rows, _, err := repos.ActivityLog.Search(context.Background(), "recover", nil, nil, nil, 1, 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("recover log rows: %+v err=%v", rows, err)
	}
}

func TestRecordActivityLogSaveError(t *testing.T) {
	repos := openTestRepos(t)
	sqlDB, err := repos.DB.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}
	err = recordRecoverActivityLog(context.Background(), repos, uuid.New(), time.Now().UTC())
	if err == nil {
		t.Fatal("expected save error after db close")
	}
}
