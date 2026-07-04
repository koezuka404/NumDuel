package db

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestOpenRedisWithAddr(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_ADDR", mr.Addr())
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("REDIS_DB", "0")

	rdb, err := OpenRedis(false)
	if err != nil {
		t.Fatalf("OpenRedis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	if err := PingRedis(context.Background(), rdb); err != nil {
		t.Fatalf("PingRedis: %v", err)
	}
}

func TestOpenRedisWithURL(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	t.Setenv("REDIS_URL", "redis://"+mr.Addr())
	t.Setenv("REDIS_ADDR", "")

	rdb, err := OpenRedis(false)
	if err != nil {
		t.Fatalf("OpenRedis url: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
}

func TestOpenRedisPingFailure(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_ADDR", "127.0.0.1:1")
	_, err := OpenRedis(false)
	if err == nil {
		t.Fatal("expected ping failure")
	}
}

func TestOpenRedisIgnoresInvalidDBNumber(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_ADDR", mr.Addr())
	t.Setenv("REDIS_DB", "not-a-number")

	rdb, err := OpenRedis(false)
	if err != nil {
		t.Fatalf("OpenRedis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
}
