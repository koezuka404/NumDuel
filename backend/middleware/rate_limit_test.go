package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/model"
)

func TestRateLimitPublicBlocksExcessLogin(t *testing.T) {
	publicRateLimiter = newRateLimiter()
	e := echo.New()
	e.POST("/api/auth/login", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, RateLimitPublic())

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.1")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestRateLimitPublicBlocksExcessRegister(t *testing.T) {
	publicRateLimiter = newRateLimiter()
	e := echo.New()
	e.POST("/api/auth/register", func(c echo.Context) error {
		return c.NoContent(http.StatusCreated)
	}, RateLimitPublic())

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", nil)
		req.Header.Set("X-Forwarded-For", "198.51.100.9")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("request %d: status %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.9")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestUserRateLimitBlocksExcessRequests(t *testing.T) {
	userRateLimiter = newRateLimiter()
	e := echo.New()
	userID := uuid.New()
	e.GET("/api/me", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			SetAuth(c, AuthInfo{UserID: userID, Role: model.RoleUser})
			return next(c)
		}
	}, UserRateLimit())

	for i := 0; i < 120; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}
