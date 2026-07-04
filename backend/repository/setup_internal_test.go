package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	_ "unsafe"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:linkname openGormFn github.com/numduel/numduel/db.openGormFn
var openGormFn func(string) (*gorm.DB, error)

//go:linkname execSQLFn github.com/numduel/numduel/db.execSQLFn
var execSQLFn func(*gorm.DB, string) error

func stubOpenSQLite(t *testing.T) {
	t.Helper()
	origOpen := openGormFn
	t.Cleanup(func() { openGormFn = origOpen })
	openGormFn = func(string) (*gorm.DB, error) {
		return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	}
}

func TestOpenDBMigrateError(t *testing.T) {
	origOpen := openGormFn
	origExec := execSQLFn
	t.Cleanup(func() {
		openGormFn = origOpen
		execSQLFn = origExec
	})

	openGormFn = func(string) (*gorm.DB, error) {
		return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	}
	execSQLFn = func(*gorm.DB, string) error {
		return errors.New("create index failed")
	}

	_, err := openDB("postgres://stub", true)
	if err == nil {
		t.Fatal("expected migrate error")
	}
}

func TestOpenDBWithoutMigrate(t *testing.T) {
	stubOpenSQLite(t)

	gdb, err := openDB("postgres://stub", false)
	if err != nil || gdb == nil {
		t.Fatalf("openDB without migrate: gdb=%v err=%v", gdb, err)
	}
}

func TestSetupRequiresDatabaseURL(t *testing.T) {
	_, err := Setup(context.Background(), SetupConfig{})
	if err == nil {
		t.Fatal("expected error for empty database URL")
	}
}

func TestSetupWithoutBackup(t *testing.T) {
	stubOpenSQLite(t)

	result, err := Setup(context.Background(), SetupConfig{DatabaseURL: "primary"})
	if err != nil {
		t.Fatalf("setup without backup: %v", err)
	}
	if result.Primary == nil || result.Repos.User == nil {
		t.Fatal("expected primary setup result")
	}
	if result.Backup != nil || result.Syncer != nil {
		t.Fatal("expected no backup configured")
	}
}

func TestSetupWithBackupStub(t *testing.T) {
	stubOpenSQLite(t)

	result, err := Setup(context.Background(), SetupConfig{
		DatabaseURL:       "primary",
		BackupDatabaseURL: "backup",
	})
	if err != nil {
		t.Fatalf("setup with backup: %v", err)
	}
	if result.Backup == nil || result.Syncer == nil {
		t.Fatal("expected backup database and syncer")
	}
}

func TestSetupBackupOpenError(t *testing.T) {
	origOpen := openGormFn
	t.Cleanup(func() { openGormFn = origOpen })

	calls := 0
	openGormFn = func(string) (*gorm.DB, error) {
		calls++
		if calls >= 2 {
			return nil, errors.New("backup unavailable")
		}
		return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	}

	_, err := Setup(context.Background(), SetupConfig{
		DatabaseURL:       "primary",
		BackupDatabaseURL: "backup",
	})
	if err == nil || !strings.Contains(err.Error(), "backup database") {
		t.Fatalf("expected backup error: %v", err)
	}
}

func TestSetupPrimaryOpenError(t *testing.T) {
	origOpen := openGormFn
	t.Cleanup(func() { openGormFn = origOpen })

	openGormFn = func(string) (*gorm.DB, error) {
		return nil, errors.New("primary unavailable")
	}

	_, err := Setup(context.Background(), SetupConfig{DatabaseURL: "primary"})
	if err == nil || !strings.Contains(err.Error(), "primary database") {
		t.Fatalf("expected primary error: %v", err)
	}
}
