// IP 単位のスライディングウィンドウで /api への過剰リクエストを 429 で拒否
package middleware

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/dto"
	"github.com/numduel/numduel/usecase"
)

type rateLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{hits: make(map[string][]time.Time)}
}

func (r *rateLimiter) allow(key string, limit int, window time.Duration) bool {
	if limit <= 0 {
		return true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	kept := r.hits[key][:0]
	for _, t := range r.hits[key] {
		if now.Sub(t) < window {
			kept = append(kept, t)
		}
	}
	if len(kept) >= limit {
		r.hits[key] = kept
		return false
	}
	r.hits[key] = append(kept, now)
	return true
}

var apiRateLimiter = newRateLimiter()

func rateLimitForPath(path string) (limit int, window time.Duration) {
	window = time.Minute
	switch path {
	case "/api/auth/login":
		return 10, window
	case "/api/auth/register":
		return 5, window
	case "/api/auth/refresh":
		return 30, window
	default:
		return 120, window
	}
}

// RateLimit は /api グループ向け IP レート制限（login/register/refresh は厳しめ）
func RateLimit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			limit, window := rateLimitForPath(c.Path())
			key := c.RealIP() + ":" + c.Path()
			if !apiRateLimiter.allow(key, limit, window) {
				return dto.WriteError(c, usecase.ErrRateLimitExceeded)
			}
			return next(c)
		}
	}
}
