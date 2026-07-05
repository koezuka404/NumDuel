package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/testutil"
)

func TestAuthRegisterLoginRefreshLogout(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodPost, "/api/auth/register", nil, map[string]string{
		"username": "alice", "email": "alice@test.local", "password": "password123",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status %d: %s", rec.Code, rec.Body.String())
	}
	registerCookies := rec.Result().Cookies()
	if len(registerCookies) == 0 {
		t.Fatal("expected cookies after register")
	}

	rec = env.do(t, http.MethodPost, "/api/auth/login", nil, map[string]string{
		"email": "alice@test.local", "password": "password123",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("login status %d: %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cookies")
	}

	rec = env.do(t, http.MethodPost, "/api/auth/refresh", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status %d: %s", rec.Code, rec.Body.String())
	}
	newCookies := rec.Result().Cookies()

	rec = env.do(t, http.MethodPost, "/api/auth/logout", newCookies, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthRegisterSetsCookies(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodPost, "/api/auth/register", nil, map[string]string{
		"username": "alice", "email": "alice@test.local", "password": "password123",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status %d", rec.Code)
	}
	var foundAccess, foundRefresh bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == middleware.AccessCookieName {
			foundAccess = true
		}
		if c.Name == middleware.RefreshCookieName {
			foundRefresh = true
		}
	}
	if !foundAccess || !foundRefresh {
		t.Fatalf("cookies missing after register: access=%v refresh=%v", foundAccess, foundRefresh)
	}
}

func TestAuthSessionWithoutCookie(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodGet, "/api/auth/session", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("session status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Data != nil {
		t.Fatalf("expected null session, got %v", body.Data)
	}
}

func TestAuthSessionWithCookie(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	loginRec := env.do(t, http.MethodPost, "/api/auth/login", nil, map[string]string{
		"email": "alice@test.local", "password": "password123",
	})
	cookies := loginRec.Result().Cookies()
	rec := env.do(t, http.MethodGet, "/api/auth/session", cookies, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("session status %d: %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"username":"alice"`)) {
		t.Fatalf("session body: %s", rec.Body.String())
	}
}

func TestAuthRegisterBadRequest(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodPost, "/api/auth/register", nil, "not-json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad register status %d", rec.Code)
	}
}

func TestAuthLoginBadRequest(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodPost, "/api/auth/login", nil, 12345)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad login status %d", rec.Code)
	}
}

func TestAuthRefreshUnauthorized(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodPost, "/api/auth/refresh", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("refresh without cookie status %d", rec.Code)
	}
}

func TestAuthLogoutUnauthorized(t *testing.T) {
	env := setupCtrlEnv(t)
	rec := env.do(t, http.MethodPost, "/api/auth/logout", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("logout unauthorized status %d", rec.Code)
	}
}

func TestAuthLoginFailure(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	rec := env.do(t, http.MethodPost, "/api/auth/login", nil, map[string]string{
		"email": "alice@test.local", "password": "wrong",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad password status %d", rec.Code)
	}
}

func TestAuthRegisterConflict(t *testing.T) {
	env := setupCtrlEnv(t)
	body := map[string]string{"username": "alice", "email": "alice@test.local", "password": "password123"}
	rec := env.do(t, http.MethodPost, "/api/auth/register", nil, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first register status %d", rec.Code)
	}
	rec = env.do(t, http.MethodPost, "/api/auth/register", nil, body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate register status %d", rec.Code)
	}
}

func TestAuthRefreshInvalidToken(t *testing.T) {
	env := setupCtrlEnv(t)
	cookies := []*http.Cookie{
		{Name: middleware.RefreshCookieName, Value: "not-a-valid-token", Path: "/api/auth/refresh"},
	}
	rec := env.do(t, http.MethodPost, "/api/auth/refresh", cookies, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid refresh status %d", rec.Code)
	}
}

func TestAuthCookiesSet(t *testing.T) {
	env := setupCtrlEnv(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	rec := env.do(t, http.MethodPost, "/api/auth/login", nil, map[string]string{
		"email": "alice@test.local", "password": "password123",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("login status %d", rec.Code)
	}
	var foundAccess, foundRefresh bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == middleware.AccessCookieName {
			foundAccess = true
		}
		if c.Name == middleware.RefreshCookieName {
			foundRefresh = true
		}
	}
	if !foundAccess || !foundRefresh {
		t.Fatalf("cookies missing: access=%v refresh=%v", foundAccess, foundRefresh)
	}
}
