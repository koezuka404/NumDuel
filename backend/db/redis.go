package db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

//OpenRedisはRedisへ接続する。required=true（本番）では未設定・到達不可時にエラーを返す。
func OpenRedis(required bool) (*redis.Client, error) {
	var (
		addr    string
		dbNum   int
		s       string
		n       int
		opts    *redis.Options
		rdb     *redis.Client
		ctxPing context.Context
		cancel  context.CancelFunc
		err     error
	)
	if raw := strings.TrimSpace(os.Getenv("REDIS_URL")); raw != "" {
		opts, err = redis.ParseURL(raw)
		if err != nil {
			return nil, fmt.Errorf("REDIS_URL: %w", err)
		}
	} else {
		addr = strings.TrimSpace(os.Getenv("REDIS_ADDR"))
		if addr == "" {
			if required {
				return nil, fmt.Errorf("REDIS_URL or REDIS_ADDR is required")
			}
			return nil, nil
		}
		dbNum = 0
		s = strings.TrimSpace(os.Getenv("REDIS_DB"))
		if s != "" {
			n, err = strconv.Atoi(s)
			if err == nil {
				dbNum = n
			}
		}
		opts = &redis.Options{
			Addr:     addr,
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       dbNum,
		}
	}
	rdb = redis.NewClient(opts)
	ctxPing, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctxPing).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return rdb, nil
}

func PingRedis(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}
	return rdb.Ping(ctx).Err()
}
