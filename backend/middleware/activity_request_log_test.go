package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/testutil"
)

func TestActivityUpdateTouchesLastActivity(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	user := testutil.CreateUser(t, repos, "act", "act@test.local", "password123")
	before, err := repos.User.FindByID(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			SetAuth(c, AuthInfo{UserID: user.ID, Role: model.RoleUser})
			return next(c)
		}
	}, ActivityUpdate(ActivityUpdateConfig{Repo: repos}))
	e.GET("/touch", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/touch", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	got, err := repos.User.FindByID(t.Context(), user.ID)
	if err != nil || !got.LastActivityAt.After(before.LastActivityAt) {
		t.Fatalf("last activity not updated: before=%v after=%v err=%v", before.LastActivityAt, got.LastActivityAt, err)
	}
}

func TestActivityUpdateSkipsWithoutAuth(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	user := testutil.CreateUser(t, repos, "noauth", "noauth@test.local", "password123")
	before, _ := repos.User.FindByID(t.Context(), user.ID)

	e := echo.New()
	e.Use(ActivityUpdate(ActivityUpdateConfig{Repo: repos}))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	got, _ := repos.User.FindByID(t.Context(), user.ID)
	if !got.LastActivityAt.Equal(before.LastActivityAt) {
		t.Fatal("should not touch without auth")
	}
}

func TestRequestLogCreatesEntry(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	user := testutil.CreateUser(t, repos, "log", "log@test.local", "password123")

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			SetAuth(c, AuthInfo{UserID: user.ID, Role: model.RoleUser})
			return next(c)
		}
	}, RequestLog(RequestLogConfig{Repo: repos}))
	e.GET("/api/me", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	uid := user.ID
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		logs, _, err := repos.ActivityLog.Search(t.Context(), "http_request", &uid, nil, nil, 1, 10)
		if err == nil && len(logs) > 0 {
			return
		}
		logs, _, err = repos.ActivityLog.Search(t.Context(), "http_request", nil, nil, nil, 1, 10)
		if err == nil && len(logs) > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("request log not created")
}

func TestShouldSkipRequestLog(t *testing.T) {
	e := echo.New()
	for _, tc := range []struct {
		method, path string
		skip         bool
	}{
		{http.MethodOptions, "/api/me", true},
		{http.MethodGet, "/health", true},
		{http.MethodGet, "/api/admin/logs", true},
		{http.MethodGet, "/api/me", false},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(tc.path)
		if got := shouldSkipRequestLog(c); got != tc.skip {
			t.Fatalf("%s %s skip=%v want %v", tc.method, tc.path, got, tc.skip)
		}
	}
}

func TestCORSEmptyOriginsIsNoOp(t *testing.T) {
	e := echo.New()
	e.Use(CORS(nil))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unexpected cors header")
	}
}
