package repository_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/numduel/numduel/repository"
)

const (
	testPrimaryDSN = "postgres://numduel:numduel@localhost:5434/numduel?sslmode=disable"
	testBackupDSN  = "postgres://numduel:numduel@localhost:5433/numduel_backup?sslmode=disable"
)

func testPrimaryURL(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	return testPrimaryDSN
}

func TestSetupRequiresDatabaseURL(t *testing.T) {
	_, err := repository.Setup(context.Background(), repository.SetupConfig{})
	if err == nil {
		t.Fatal("expected error for empty database URL")
	}
}

func TestSetupInvalidDSN(t *testing.T) {
	_, err := repository.Setup(context.Background(), repository.SetupConfig{
		DatabaseURL: "postgres://invalid:5432/nope",
		Migrate:     false,
	})
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
}

func TestSetupPostgresPrimaryOnly(t *testing.T) {
	result, err := repository.Setup(context.Background(), repository.SetupConfig{
		DatabaseURL: testPrimaryURL(t),
		Migrate:     false,
	})
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if result.Primary == nil || result.Repos.User == nil {
		t.Fatal("expected primary setup result")
	}
	if result.Backup != nil || result.Syncer != nil {
		t.Fatal("expected no backup configured")
	}
}

func TestSetupPostgresWithMigrate(t *testing.T) {
	result, err := repository.Setup(context.Background(), repository.SetupConfig{
		DatabaseURL: testPrimaryURL(t),
		Migrate:     true,
	})
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if result.Primary == nil {
		t.Fatal("expected migrated primary database")
	}
}

func TestSetupInvalidBackupDSN(t *testing.T) {
	if _, err := repository.Setup(context.Background(), repository.SetupConfig{
		DatabaseURL: testPrimaryURL(t),
		Migrate:     false,
	}); err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}

	_, err := repository.Setup(context.Background(), repository.SetupConfig{
		DatabaseURL:       testPrimaryURL(t),
		BackupDatabaseURL: "postgres://invalid:5432/nope",
		Migrate:           false,
	})
	if err == nil {
		t.Fatal("expected backup setup error")
	}
	if !strings.Contains(err.Error(), "backup database") {
		t.Fatalf("expected backup database error, got: %v", err)
	}
}

func TestSetupWithBackup(t *testing.T) {
	backupDSN := os.Getenv("TEST_BACKUP_DATABASE_URL")
	if backupDSN == "" {
		backupDSN = testBackupDSN
	}
	result, err := repository.Setup(context.Background(), repository.SetupConfig{
		DatabaseURL:       testPrimaryURL(t),
		BackupDatabaseURL: backupDSN,
		Migrate:           false,
	})
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if result.Backup == nil || result.Syncer == nil {
		t.Fatal("expected backup database and syncer")
	}
}

func TestSetupUserSearchPostgres(t *testing.T) {
	result, err := repository.Setup(context.Background(), repository.SetupConfig{
		DatabaseURL: testPrimaryURL(t),
		Migrate:     false,
	})
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	username := "srch_" + suffix
	user := newUser(username, username+"@test.local")
	if err := result.Repos.User.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	found, total, err := result.Repos.User.Search(ctx, "srch_"+suffix[:4], 1, 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if total == 0 || len(found) == 0 {
		t.Fatalf("search results: total=%d len=%d", total, len(found))
	}
}
