package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/testutil"
)

// セキュリティ: 認可 middleware
func TestAuthRejectsRevokedToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	revoker := testutil.NewMemJWTRevoker()
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := jwtSvc.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := revoker.Revoke(context.Background(), claims.JTI, time.Hour); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	e := echo.New()
	e.GET("/protected", func(c echo.Context) error { return c.NoContent(http.StatusOK) },
		middleware.Auth(middleware.AuthConfig{JWT: jwtSvc, Revoker: revoker, Repo: repos}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: middleware.AccessCookieName, Value: token})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked token status %d", rec.Code)
	}
}

func TestAuthRejectsBearerHeaderWithoutCookie(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	e := echo.New()
	e.GET("/protected", func(c echo.Context) error { return c.NoContent(http.StatusOK) },
		middleware.Auth(middleware.AuthConfig{JWT: jwtSvc, Repo: repos}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bearer without cookie status %d", rec.Code)
	}
}

func TestAuthRejectsDeletedUser(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	now := time.Now().UTC()
	user.DeletedAt = &now
	if err := repos.User.Update(context.Background(), user); err != nil {
		t.Fatalf("delete: %v", err)
	}

	e := echo.New()
	e.GET("/protected", func(c echo.Context) error { return c.NoContent(http.StatusOK) },
		middleware.Auth(middleware.AuthConfig{JWT: jwtSvc, Repo: repos}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: middleware.AccessCookieName, Value: token})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("deleted user status %d", rec.Code)
	}
}

func TestAdminMiddlewareRejectsNonMaster(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	e := echo.New()
	e.GET("/admin", func(c echo.Context) error { return c.NoContent(http.StatusOK) },
		middleware.Auth(middleware.AuthConfig{JWT: jwtSvc, Repo: repos}), middleware.Admin())

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: middleware.AccessCookieName, Value: token})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-master admin status %d", rec.Code)
	}
}

func TestAuthRejectsTamperedToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	token, err := jwtSvc.Issue(user.ID, user.Role, time.Now().UTC())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	tampered := token[:len(token)-1] + "X"

	e := echo.New()
	e.GET("/protected", func(c echo.Context) error { return c.NoContent(http.StatusOK) },
		middleware.Auth(middleware.AuthConfig{JWT: jwtSvc, Repo: repos}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: middleware.AccessCookieName, Value: tampered})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("tampered token status %d", rec.Code)
	}
}

func TestAuthRejectsForceLogoutToken(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	force := testutil.NewMemForceLogout()
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	issuedAt := time.Now().UTC().Add(-30 * time.Minute)
	token, err := jwtSvc.Issue(user.ID, user.Role, issuedAt)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if err := force.SetForceLogoutBefore(context.Background(), user.ID, time.Now().UTC()); err != nil {
		t.Fatalf("force logout: %v", err)
	}

	e := echo.New()
	e.GET("/protected", func(c echo.Context) error { return c.NoContent(http.StatusOK) },
		middleware.Auth(middleware.AuthConfig{JWT: jwtSvc, ForceLogout: force, Repo: repos}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: middleware.AccessCookieName, Value: token})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("force logout token status %d", rec.Code)
	}
}
