//IP/ユーザー単位のスライディングウィンドウで/apiへの過剰リクエストを429で拒否
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

var publicRateLimiter = newRateLimiter()
var userRateLimiter = newRateLimiter()

func publicRateLimitForPath(path string) (limit int, window time.Duration, apply bool) {
	window = time.Minute
	switch path {
	case "/api/auth/login":
		return 10, window, true
	case "/api/auth/register":
		return 5, window, true
	case "/api/auth/refresh":
		return 30, window, true
	default:
		return 0, window, false
	}
}

//RateLimitPublicはlogin/register/refresh向けIPレート制限（§11.8）
func RateLimitPublic() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			limit, window, apply := publicRateLimitForPath(c.Path())
			if !apply {
				return next(c)
			}
			key := c.RealIP() + ":" + c.Path()
			if !publicRateLimiter.allow(key, limit, window) {
				return dto.WriteError(c, usecase.ErrRateLimitExceeded)
			}
			return next(c)
		}
	}
}

//UserRateLimitは認証済みAPI向けユーザー単位120回/分（§11.8）。Authの後に適用する。
func UserRateLimit() echo.MiddlewareFunc {
	const limit = 120
	window := time.Minute
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, ok := AuthFrom(c)
			if !ok {
				return next(c)
			}
			key := auth.UserID.String()
			if !userRateLimiter.allow(key, limit, window) {
				return dto.WriteError(c, usecase.ErrRateLimitExceeded)
			}
			return next(c)
		}
	}
}
