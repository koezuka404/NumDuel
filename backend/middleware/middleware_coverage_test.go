package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
	"github.com/numduel/numduel/testutil"
)

type errRevoker struct{}

func (errRevoker) Revoke(context.Context, string, time.Duration) error { return nil }

func (errRevoker) IsRevoked(context.Context, string) (bool, error) {
	return false, context.Canceled
}

type errForceLogoutStore struct{}

func (errForceLogoutStore) GetForceLogoutBefore(context.Context, uuid.UUID) (time.Time, error) {
	return time.Time{}, context.Canceled
}

func (errForceLogoutStore) SetForceLogoutBefore(context.Context, uuid.UUID, time.Time) error {
	return nil
}

type nilUserFindRepo struct {
	repository.IUserRepo
}

func (nilUserFindRepo) FindByID(context.Context, uuid.UUID) (*model.User, error) {
	return nil, nil
}

type errUserFindRepo struct {
	repository.IUserRepo
}

func (errUserFindRepo) FindByID(context.Context, uuid.UUID) (*model.User, error) {
	return nil, context.Canceled
}

func authEcho(t *testing.T, cfg AuthConfig) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.GET("/protected", func(c echo.Context) error { return c.NoContent(http.StatusOK) }, Auth(cfg))
	return e
}

func authRequest(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: token})
	}
	return req
}

func TestAdminAllowsMaster(t *testing.T) {
	e := echo.New()
	h := Admin()(func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	SetAuth(c, AuthInfo{UserID: uuid.New(), Role: model.RoleMaster})
	if err := h(c); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestRateLimitPublicRefreshPath(t *testing.T) {
	limit, _, apply := publicRateLimitForPath("/api/auth/refresh")
	if !apply || limit != 30 {
		t.Fatalf("refresh limit=%d apply=%v", limit, apply)
	}
}

func TestUserRateLimitSkipsWithoutAuth(t *testing.T) {
	e := echo.New()
	e.Use(UserRateLimit())
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestRateLimiterDeniesWhenOverLimit(t *testing.T) {
	rl := newRateLimiter()
	key := "k"
	if !rl.allow(key, 1, time.Minute) || rl.allow(key, 1, time.Minute) {
		t.Fatal("second request should be denied")
	}
}

func TestAuthAllowsValidToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "valid", "valid@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	e := authEcho(t, AuthConfig{JWT: jwtSvc, Repo: repos})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authRequest(token))
	if rec.Code != http.StatusOK {
		t.Fatalf("valid token status %d", rec.Code)
	}
}

func TestAuthRejectsEmptyCookieValue(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}

	e := authEcho(t, AuthConfig{JWT: jwtSvc, Repo: repos})
	req := authRequest("")
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: ""})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("empty cookie status %d", rec.Code)
	}
}

func TestAuthRevokerCheckError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "rev", "rev@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	e := authEcho(t, AuthConfig{JWT: jwtSvc, Revoker: errRevoker{}, Repo: repos})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authRequest(token))
	if rec.Code != http.StatusInternalServerError && rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoker error status %d", rec.Code)
	}
}

func TestAuthForceLogoutCheckError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "force", "force@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	e := authEcho(t, AuthConfig{JWT: jwtSvc, ForceLogout: errForceLogoutStore{}, Repo: repos})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authRequest(token))
	if rec.Code != http.StatusInternalServerError && rec.Code != http.StatusUnauthorized {
		t.Fatalf("force logout error status %d", rec.Code)
	}
}

func TestAuthRejectsMissingUserRecord(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "gone", "gone@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	repos.User = nilUserFindRepo{IUserRepo: repos.User}
	e := authEcho(t, AuthConfig{JWT: jwtSvc, Repo: repos})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authRequest(token))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing user status %d", rec.Code)
	}
}

func TestAuthUserFindError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "find", "find@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	repos.User = errUserFindRepo{IUserRepo: repos.User}
	e := authEcho(t, AuthConfig{JWT: jwtSvc, Repo: repos})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, authRequest(token))
	if rec.Code != http.StatusInternalServerError && rec.Code != http.StatusUnauthorized {
		t.Fatalf("find user error status %d", rec.Code)
	}
}

func TestRateLimitPublicSkipsUnlistedPath(t *testing.T) {
	publicRateLimiter = newRateLimiter()
	e := echo.New()
	e.GET("/api/me", func(c echo.Context) error { return c.NoContent(http.StatusOK) }, RateLimitPublic())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestRateLimitPublicBlocksExcessRefresh(t *testing.T) {
	publicRateLimiter = newRateLimiter()
	e := echo.New()
	e.POST("/api/auth/refresh", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, RateLimitPublic())

	ip := "203.0.113.44"
	for i := 0; i < 30; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		req.Header.Set("X-Forwarded-For", ip)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.Header.Set("X-Forwarded-For", ip)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestRequestLogSkipsOptionsThroughMiddleware(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	e := echo.New()
	e.Use(RequestLog(RequestLogConfig{Repo: repos}))
	e.OPTIONS("/api/me", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/api/me", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}

	time.Sleep(100 * time.Millisecond)
	logs, _, err := repos.ActivityLog.Search(t.Context(), "http_request", nil, nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(logs) != 0 {
		t.Fatal("OPTIONS should not create request log")
	}
}

func TestRequestLogSkipsHealthThroughMiddleware(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	e := echo.New()
	e.Use(RequestLog(RequestLogConfig{Repo: repos}))
	e.GET("/health", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	time.Sleep(100 * time.Millisecond)

	logs, _, err := repos.ActivityLog.Search(t.Context(), "http_request", nil, nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(logs) != 0 {
		t.Fatal("/health should not create request log")
	}
}

func TestRequestLogNilUserRepo(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	cfg := RequestLogConfig{Repo: repository.Repos{ActivityLog: repos.ActivityLog}}

	e := echo.New()
	e.Use(RequestLog(cfg))
	e.GET("/api/x", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/x", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestRequestLogDefaultStatusWhenZero(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	e := echo.New()
	h := RequestLog(RequestLogConfig{Repo: repos})(func(c echo.Context) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/api/zero", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/zero")
	if err := h(c); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if c.Response().Status != 0 {
		t.Fatalf("expected zero status before write, got %d", c.Response().Status)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		logs, _, err := repos.ActivityLog.Search(t.Context(), "http_request", nil, nil, nil, 1, 10)
		if err == nil && len(logs) > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("request log not created for zero status response")
}

func TestRequestLogCreateActivityLogError(t *testing.T) {
	gdb, repos := testutil.OpenSQLiteDB(t)
	e := echo.New()
	e.Use(RequestLog(RequestLogConfig{Repo: repos}))
	e.GET("/api/fail", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/fail", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
}

func TestRequestLogMarshalError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	orig := marshalRequestLogDetail
	t.Cleanup(func() { marshalRequestLogDetail = orig })
	marshalRequestLogDetail = func(any) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}

	e := echo.New()
	h := RequestLog(RequestLogConfig{Repo: repos})(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/me")
	if err := h(c); err != nil {
		t.Fatalf("handler: %v", err)
	}
}

func TestRequestLogReturnsHandlerError(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	e := echo.New()
	e.Use(RequestLog(RequestLogConfig{Repo: repos}))
	e.GET("/api/bad", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "bad")
	})

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/bad", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}
