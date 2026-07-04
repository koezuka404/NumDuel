package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestOpenEmptyDSN(t *testing.T) {
	_, err := Open("")
	if err == nil || err.Error() != "DATABASE_URL is empty" {
		t.Fatalf("Open: %v", err)
	}
}

func TestOpenSuccessWithStub(t *testing.T) {
	orig := openGormFn
	t.Cleanup(func() { openGormFn = orig })
	openGormFn = func(string) (*gorm.DB, error) {
		return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	}
	gdb, err := Open("postgres://stub")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	sqlDB, err := SQLDB(gdb)
	if err != nil || sqlDB == nil {
		t.Fatalf("SQLDB: %v", err)
	}
	if err := Ping(context.Background(), gdb); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpenGormError(t *testing.T) {
	orig := openGormFn
	t.Cleanup(func() { openGormFn = orig })
	openGormFn = func(string) (*gorm.DB, error) {
		return nil, errors.New("open failed")
	}
	_, err := Open("postgres://bad")
	if err == nil {
		t.Fatal("expected open error")
	}
}

func TestSQLDBFromGormNil(t *testing.T) {
	_, err := sqlDBFromGorm(nil)
	if err == nil {
		t.Fatal("expected error for nil gorm db")
	}
}

func TestIntFromEnv(t *testing.T) {
	t.Setenv("DB_MAX_IDLE_CONNS", "10")
	if intFromEnv("DB_MAX_IDLE_CONNS", 20) != 10 {
		t.Fatal("expected parsed int")
	}
	t.Setenv("DB_MAX_IDLE_CONNS", "bad")
	if intFromEnv("DB_MAX_IDLE_CONNS", 20) != 20 {
		t.Fatal("expected fallback")
	}
	t.Setenv("DB_MAX_IDLE_CONNS", "0")
	if intFromEnv("DB_MAX_IDLE_CONNS", 20) != 20 {
		t.Fatal("expected fallback for zero")
	}
	if intFromEnv("UNSET_VAR", 7) != 7 {
		t.Fatal("expected default")
	}
}

func TestDurationFromEnv(t *testing.T) {
	t.Setenv("DB_CONN_MAX_LIFETIME", "10m")
	if durationFromEnv("DB_CONN_MAX_LIFETIME", time.Minute) != 10*time.Minute {
		t.Fatal("expected parsed duration")
	}
	t.Setenv("DB_CONN_MAX_LIFETIME", "bad")
	if durationFromEnv("DB_CONN_MAX_LIFETIME", time.Minute) != time.Minute {
		t.Fatal("expected fallback")
	}
}

func TestConfigureSQLDB(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Setenv("DB_MAX_IDLE_CONNS", "5")
	t.Setenv("DB_MAX_OPEN_CONNS", "10")
	t.Setenv("DB_CONN_MAX_LIFETIME", "1h")
	t.Setenv("DB_CONN_MAX_IDLE_TIME", "30s")
	configureSQLDB(sqlDB)
}

func TestMigrate(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
}

func TestExecSQL(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := execSQL(gdb, "SELECT 1"); err != nil {
		t.Fatalf("execSQL: %v", err)
	}
}
