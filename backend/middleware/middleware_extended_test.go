package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/model"
)

func TestAuthContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	info := AuthInfo{UserID: uuid.New(), Role: model.RoleUser, JTI: "jti", ExpiresAt: time.Now().UTC()}
	SetAuth(c, info)
	got, ok := AuthFrom(c)
	if !ok || got.UserID != info.UserID || got.JTI != "jti" {
		t.Fatalf("AuthFrom: %+v ok=%v", got, ok)
	}
}

func TestAdminMiddlewareUnauthorized(t *testing.T) {
	e := echo.New()
	e.GET("/admin", func(c echo.Context) error { return c.NoContent(http.StatusOK) }, Admin())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestCORSAllowsOrigin(t *testing.T) {
	e := echo.New()
	e.Use(CORS([]string{"http://allowed.test"}))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://allowed.test")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://allowed.test" {
		t.Fatalf("cors header missing")
	}
}

func TestRecoverMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(Recover())
	e.GET("/panic", func(c echo.Context) error { panic("boom") })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestPublicRateLimitForPath(t *testing.T) {
	if limit, _, apply := publicRateLimitForPath("/api/auth/login"); !apply || limit != 10 {
		t.Fatalf("login limit=%d apply=%v", limit, apply)
	}
	if _, _, apply := publicRateLimitForPath("/api/me"); apply {
		t.Fatal("me should not rate limit public")
	}
}

func TestRateLimiterAllowZeroLimit(t *testing.T) {
	rl := newRateLimiter()
	if !rl.allow("k", 0, time.Minute) {
		t.Fatal("zero limit should allow")
	}
}
