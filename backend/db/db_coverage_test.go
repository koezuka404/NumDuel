package db

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestOpenPostgresInvalidDSN(t *testing.T) {
	_, err := openPostgres("not-a-valid-dsn")
	if err == nil {
		t.Fatal("expected openPostgres error")
	}
}

func TestOpenSQLDBFromGormError(t *testing.T) {
	orig := openGormFn
	t.Cleanup(func() { openGormFn = orig })
	openGormFn = func(string) (*gorm.DB, error) {
		return nil, nil
	}
	_, err := Open("postgres://stub")
	if err == nil {
		t.Fatal("expected sql handle error")
	}
}

func TestPingNilGormDB(t *testing.T) {
	if err := Ping(context.Background(), nil); err == nil {
		t.Fatal("expected ping error for nil gorm db")
	}
}

func TestMigrateAutoMigrateError(t *testing.T) {
	orig := autoMigrateFn
	t.Cleanup(func() { autoMigrateFn = orig })
	autoMigrateFn = func(*gorm.DB) error {
		return errors.New("migrate failed")
	}
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := Migrate(gdb); err == nil {
		t.Fatal("expected auto migrate error")
	}
}

func TestMigrateIndexError(t *testing.T) {
	orig := execSQLFn
	t.Cleanup(func() { execSQLFn = orig })
	execSQLFn = func(*gorm.DB, string) error {
		return errors.New("index failed")
	}
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := Migrate(gdb); err == nil {
		t.Fatal("expected index error")
	}
}

func TestOpenRedisInvalidURL(t *testing.T) {
	t.Setenv("REDIS_URL", "://bad")
	t.Setenv("REDIS_ADDR", "")
	_, err := OpenRedis(false)
	if err == nil {
		t.Fatal("expected REDIS_URL parse error")
	}
}

func TestOpenRedisRequiredWithoutConfig(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_ADDR", "")
	_, err := OpenRedis(true)
	if err == nil {
		t.Fatal("expected required redis error")
	}
}

func TestOpenRedisOptionalWithoutConfig(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_ADDR", "")
	rdb, err := OpenRedis(false)
	if err != nil {
		t.Fatalf("OpenRedis: %v", err)
	}
	if rdb != nil {
		t.Fatal("expected nil client when redis is optional and unset")
	}
}

func TestOpenRedisWithValidDBNumber(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_ADDR", mr.Addr())
	t.Setenv("REDIS_DB", "2")

	rdb, err := OpenRedis(false)
	if err != nil {
		t.Fatalf("OpenRedis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
}

func TestPingRedisNil(t *testing.T) {
	if err := PingRedis(context.Background(), nil); err != nil {
		t.Fatalf("PingRedis(nil): %v", err)
	}
}

func TestPingRedisError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	mr.Close()
	t.Cleanup(func() { _ = rdb.Close() })

	if err := PingRedis(context.Background(), rdb); err == nil {
		t.Fatal("expected ping error on closed redis")
	}
}
