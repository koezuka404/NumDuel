package db

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func OpenRedis() *redis.Client {
	var (
		addr    string
		dbNum   int
		s       string
		n       int
		rdb     *redis.Client
		ctxPing context.Context
		cancel  context.CancelFunc
		err     error
	)
	addr = strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr == "" {
		return nil
	}
	dbNum = 0
	s = strings.TrimSpace(os.Getenv("REDIS_DB"))
	if s != "" {
		n, err = strconv.Atoi(s)
		if err == nil {
			dbNum = n
		}
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbNum,
	})
	ctxPing, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = rdb.Ping(ctxPing).Err()
	return rdb
}

func PingRedis(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}
	return rdb.Ping(ctx).Err()
}
